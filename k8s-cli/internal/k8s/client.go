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

// Client обертка для Kubernetes клиента
type Client struct {
	clientset     *kubernetes.Clientset
	dynamicClient dynamic.Interface
	config        clientcmd.ClientConfig
}

// NewClient создает новый Kubernetes клиент
func NewClient(kubeconfigPath string) (*Client, error) {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{},
	)

	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("ошибка создания конфигурации: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания dynamic клиента: %w", err)
	}

	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		config:        config,
	}, nil
}

// GetClientset возвращает Kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// GetDynamicClient возвращает dynamic клиент
func (c *Client) GetDynamicClient() dynamic.Interface {
	return c.dynamicClient
}

// GetCurrentContext возвращает текущий контекст
func (c *Client) GetCurrentContext() (string, error) {
	rawConfig, err := c.config.RawConfig()
	if err != nil {
		return "", err
	}
	return rawConfig.CurrentContext, nil
}

// GetContexts возвращает список всех контекстов
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

// SetContext переключает контекст
func (c *Client) SetContext(contextName, kubeconfigPath string) error {
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("ошибка загрузки kubeconfig: %w", err)
	}

	if _, exists := config.Contexts[contextName]; !exists {
		return fmt.Errorf("контекст '%s' не найден", contextName)
	}

	config.CurrentContext = contextName

	return clientcmd.WriteToFile(*config, kubeconfigPath)
}

// TestConnection проверяет соединение с кластером
func (c *Client) TestConnection() error {
	_, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("не удается подключиться к кластеру: %w", err)
	}
	return nil
}

// CreateFromYAML создает ресурс из YAML
func (c *Client) CreateFromYAML(yamlData []byte, namespace string) error {
	// Декодируем YAML в unstructured объект
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(yamlData)), 4096)

	var obj unstructured.Unstructured
	if err := decoder.Decode(&obj); err != nil {
		return fmt.Errorf("ошибка декодирования YAML: %w", err)
	}

	// Получаем GVK из объекта
	gvk := obj.GroupVersionKind()

	// Устанавливаем namespace если он не указан и это namespaced ресурс
	if obj.GetNamespace() == "" && namespace != "" && !isClusterScoped(gvk.Kind) {
		obj.SetNamespace(namespace)
	}

	// Создаем ресурс используя dynamic клиент
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
		return fmt.Errorf("ошибка создания ресурса: %w", err)
	}

	return nil
}

// isClusterScoped проверяет является ли ресурс cluster-scoped
func isClusterScoped(kind string) bool {
	clusterScopedResources := []string{
		"Namespace",
		"Node",
		"PersistentVolume",
		"ClusterRole",
		"ClusterRoleBinding",
		"StorageClass",
	}

	for _, resource := range clusterScopedResources {
		if kind == resource {
			return true
		}
	}
	return false
}

// getResourceName возвращает имя ресурса по Kind
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
	default:
		// Простая конвертация - добавляем 's'
		return strings.ToLower(kind) + "s"
	}
}
