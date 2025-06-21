package k8s

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wrapper for Kubernetes client
type Client struct {
	clientset     *kubernetes.Clientset
	dynamicClient dynamic.Interface
	config        clientcmd.ClientConfig
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfigPath string) (*Client, error) {
	// If kubeconfigPath is empty, use default kubeconfig location
	if kubeconfigPath == "" {
		kubeconfigPath = clientcmd.RecommendedHomeFile
	}

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{},
	)

	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error creating configuration: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic client: %w", err)
	}

	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		config:        config,
	}, nil
}

// GetClientset returns the Kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// GetDynamicClient returns the dynamic client
func (c *Client) GetDynamicClient() dynamic.Interface {
	return c.dynamicClient
}

// GetCurrentContext returns the current context
func (c *Client) GetCurrentContext() (string, error) {
	rawConfig, err := c.config.RawConfig()
	if err != nil {
		return "", err
	}
	return rawConfig.CurrentContext, nil
}

// GetContexts returns a list of all contexts
func (c *Client) GetContexts() ([]string, error) {
	rawConfig, err := c.config.RawConfig()
	if err != nil {
		return nil, err
	}

	var contexts []string
	for name := range rawConfig.Contexts {
		contexts = append(contexts, name)
	}
	return contexts, nil
}

// SetContext switches the context
func (c *Client) SetContext(contextName, kubeconfigPath string) error {
	if kubeconfigPath == "" {
		kubeconfigPath = clientcmd.RecommendedHomeFile
	}

	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("error loading kubeconfig: %w", err)
	}

	if _, exists := config.Contexts[contextName]; !exists {
		return fmt.Errorf("context '%s' not found", contextName)
	}

	config.CurrentContext = contextName

	return clientcmd.WriteToFile(*config, kubeconfigPath)
}

// TestConnection tests the connection to the cluster
func (c *Client) TestConnection() error {
	_, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("unable to connect to cluster: %w", err)
	}
	return nil
}

// CreateFromYAML creates a resource from YAML
func (c *Client) CreateFromYAML(yamlData []byte, namespace string) error {
	// Decode YAML into unstructured object
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(yamlData)), 4096)

	var obj unstructured.Unstructured
	if err := decoder.Decode(&obj); err != nil {
		return fmt.Errorf("error decoding YAML: %w", err)
	}

	// Get GVK from object
	gvk := obj.GroupVersionKind()

	// Set namespace if not specified and this is a namespaced resource
	if obj.GetNamespace() == "" && namespace != "" && !isClusterScoped(gvk.Kind) {
		obj.SetNamespace(namespace)
	}

	// Create resource using dynamic client
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: getResourceName(gvk.Kind),
	}

	var err error
	if isClusterScoped(gvk.Kind) {
		_, err = c.dynamicClient.Resource(gvr).Create(
			context.TODO(),
			&obj,
			metav1.CreateOptions{},
		)
	} else {
		_, err = c.dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Create(
			context.TODO(),
			&obj,
			metav1.CreateOptions{},
		)
	}

	if err != nil {
		return fmt.Errorf("error creating resource: %w", err)
	}

	return nil
}

// DeleteFromYAML deletes a resource from YAML
func (c *Client) DeleteFromYAML(yamlData []byte, namespace string) error {
	// Decode YAML into unstructured object
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(yamlData)), 4096)

	var obj unstructured.Unstructured
	if err := decoder.Decode(&obj); err != nil {
		return fmt.Errorf("error decoding YAML: %w", err)
	}

	// Get GVK from object
	gvk := obj.GroupVersionKind()

	// Set namespace if not specified and this is a namespaced resource
	if obj.GetNamespace() == "" && namespace != "" && !isClusterScoped(gvk.Kind) {
		obj.SetNamespace(namespace)
	}

	// Delete resource using dynamic client
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: getResourceName(gvk.Kind),
	}

	var err error
	if isClusterScoped(gvk.Kind) {
		err = c.dynamicClient.Resource(gvr).Delete(
			context.TODO(),
			obj.GetName(),
			metav1.DeleteOptions{},
		)
	} else {
		err = c.dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Delete(
			context.TODO(),
			obj.GetName(),
			metav1.DeleteOptions{},
		)
	}

	if err != nil {
		return fmt.Errorf("error deleting resource: %w", err)
	}

	return nil
}

// ListDeployments lists deployments in the specified namespace (Step 6 requirement)
func (c *Client) ListDeployments(namespace string) error {
	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		return fmt.Errorf("error listing deployments: %w", err)
	}

	fmt.Printf("Deployments in namespace '%s':\n", namespace)
	for _, deployment := range deployments.Items {
		replicas := int32(0)
		if deployment.Spec.Replicas != nil {
			replicas = *deployment.Spec.Replicas
		}
		fmt.Printf("  %s - Ready: %d/%d, Available: %d\n",
			deployment.Name,
			deployment.Status.ReadyReplicas,
			replicas,
			deployment.Status.AvailableReplicas,
		)
	}

	return nil
}

// isClusterScoped checks if a resource is cluster-scoped
func isClusterScoped(kind string) bool {
	clusterScopedResources := []string{
		"Namespace",
		"Node",
		"PersistentVolume",
		"ClusterRole",
		"ClusterRoleBinding",
		"StorageClass",
		"CustomResourceDefinition",
		"ValidatingAdmissionWebhook",
		"MutatingAdmissionWebhook",
	}

	for _, resource := range clusterScopedResources {
		if kind == resource {
			return true
		}
	}
	return false
}

// getResourceName returns the resource name by Kind
func getResourceName(kind string) string {
	switch kind {
	case "Pod":
		return "pods"
	case "Deployment":
		return "deployments"
	case "Service":
		return "services"
	case "ConfigMap":
		return "configmaps"
	case "Secret":
		return "secrets"
	case "Namespace":
		return "namespaces"
	case "Ingress":
		return "ingresses"
	case "PersistentVolume":
		return "persistentvolumes"
	case "PersistentVolumeClaim":
		return "persistentvolumeclaims"
	case "ServiceAccount":
		return "serviceaccounts"
	case "Role":
		return "roles"
	case "RoleBinding":
		return "rolebindings"
	case "ClusterRole":
		return "clusterroles"
	case "ClusterRoleBinding":
		return "clusterrolebindings"
	case "DaemonSet":
		return "daemonsets"
	case "StatefulSet":
		return "statefulsets"
	case "ReplicaSet":
		return "replicasets"
	case "Job":
		return "jobs"
	case "CronJob":
		return "cronjob"
	default:
		return strings.ToLower(kind) + "s"
	}
}
