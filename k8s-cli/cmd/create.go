package cmd

import (
	"context"
	"fmt"
	"k8s-cli/internal/k8s"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Kubernetes resources imperatively",
	Long:  "Create Kubernetes resources imperatively (without YAML files)",
}

// createDeploymentCmd creates a deployment imperatively
var createDeploymentCmd = &cobra.Command{
	Use:   "deployment <name>",
	Short: "Create a deployment",
	Long:  "Create a deployment imperatively with specified image",
	Args:  cobra.ExactArgs(1),
	Example: `  # Create a deployment
  k8s-cli create deployment nginx --image=nginx:1.20

  # Create deployment with specific replica count
  k8s-cli create deployment app --image=gcr.io/kuber-351315/week-3:v1.0.0 --replicas=3

  # Create deployment in specific namespace
  k8s-cli create deployment demo2 --image=gcr.io/kuber-351315/week-3:v1.0.0 -n my-namespace`,
	RunE: runCreateDeployment,
}

// createPodCmd creates a pod imperatively
var createPodCmd = &cobra.Command{
	Use:   "pod <name>",
	Short: "Create a pod",
	Long:  "Create a pod imperatively with specified image",
	Args:  cobra.ExactArgs(1),
	Example: `  # Create a pod
  k8s-cli create pod nginx --image=nginx:1.20

  # Create pod in specific namespace
  k8s-cli create pod test-pod --image=gcr.io/kuber-351315/week-3:v1.0.0 -n my-namespace`,
	RunE: runCreatePod,
}

// createServiceCmd creates a service imperatively
var createServiceCmd = &cobra.Command{
	Use:   "service <name>",
	Short: "Create a service",
	Long:  "Create a service imperatively",
	Args:  cobra.ExactArgs(1),
	Example: `  # Create a ClusterIP service
  k8s-cli create service my-service --port=80 --target-port=8080

  # Create a NodePort service
  k8s-cli create service my-service --port=80 --target-port=8080 --type=NodePort

  # Create service with selector
  k8s-cli create service my-service --port=80 --selector=app=nginx`,
	RunE: runCreateService,
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createDeploymentCmd)
	createCmd.AddCommand(createPodCmd)
	createCmd.AddCommand(createServiceCmd)

	// Flags for deployment
	createDeploymentCmd.Flags().String("image", "", "Container image to use (required)")
	createDeploymentCmd.Flags().Int32("replicas", 1, "Number of replicas")
	createDeploymentCmd.Flags().Int32("port", 0, "Container port to expose")
	createDeploymentCmd.MarkFlagRequired("image")

	// Flags for pod
	createPodCmd.Flags().String("image", "", "Container image to use (required)")
	createPodCmd.Flags().Int32("port", 0, "Container port to expose")
	createPodCmd.MarkFlagRequired("image")

	// Flags for service
	createServiceCmd.Flags().Int32("port", 80, "Service port")
	createServiceCmd.Flags().Int32("target-port", 0, "Target port (defaults to port)")
	createServiceCmd.Flags().String("type", "ClusterIP", "Service type (ClusterIP, NodePort, LoadBalancer)")
	createServiceCmd.Flags().String("selector", "", "Selector for service (e.g., app=nginx)")
}

func runCreateDeployment(cmd *cobra.Command, args []string) error {
	deploymentName := args[0]
	image, _ := cmd.Flags().GetString("image")
	replicas, _ := cmd.Flags().GetInt32("replicas")
	port, _ := cmd.Flags().GetInt32("port")
	namespace := viper.GetString("namespace")

	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	// Create deployment object
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": deploymentName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  deploymentName,
							Image: image,
						},
					},
				},
			},
		},
	}

	// Add port if specified
	if port > 0 {
		deployment.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{
				ContainerPort: port,
			},
		}
	}

	// Create the deployment
	_, err = client.GetClientset().AppsV1().Deployments(namespace).Create(
		context.TODO(),
		deployment,
		metav1.CreateOptions{},
	)
	if err != nil {
		return fmt.Errorf("error creating deployment: %w", err)
	}

	fmt.Printf("✅ Deployment '%s' created successfully in namespace '%s'\n", deploymentName, namespace)
	fmt.Printf("   Image: %s\n", image)
	fmt.Printf("   Replicas: %d\n", replicas)
	if port > 0 {
		fmt.Printf("   Port: %d\n", port)
	}

	return nil
}

func runCreatePod(cmd *cobra.Command, args []string) error {
	podName := args[0]
	image, _ := cmd.Flags().GetString("image")
	port, _ := cmd.Flags().GetInt32("port")
	namespace := viper.GetString("namespace")

	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	// Create pod object
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": podName,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  podName,
					Image: image,
				},
			},
		},
	}

	// Add port if specified
	if port > 0 {
		pod.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{
				ContainerPort: port,
			},
		}
	}

	// Create the pod
	_, err = client.GetClientset().CoreV1().Pods(namespace).Create(
		context.TODO(),
		pod,
		metav1.CreateOptions{},
	)
	if err != nil {
		return fmt.Errorf("error creating pod: %w", err)
	}

	fmt.Printf("✅ Pod '%s' created successfully in namespace '%s'\n", podName, namespace)
	fmt.Printf("   Image: %s\n", image)
	if port > 0 {
		fmt.Printf("   Port: %d\n", port)
	}

	return nil
}

func runCreateService(cmd *cobra.Command, args []string) error {
	serviceName := args[0]
	port, _ := cmd.Flags().GetInt32("port")
	targetPort, _ := cmd.Flags().GetInt32("target-port")
	serviceType, _ := cmd.Flags().GetString("type")
	selector, _ := cmd.Flags().GetString("selector")
	namespace := viper.GetString("namespace")

	// Default target port to port if not specified
	if targetPort == 0 {
		targetPort = port
	}

	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	// Parse selector
	selectorMap := make(map[string]string)
	if selector != "" {
		// Parse selector string like "app=nginx,version=v1"
		pairs := []string{selector} // Simple implementation for single selector
		for _, pair := range pairs {
			if parts := splitKeyValue(pair); len(parts) == 2 {
				selectorMap[parts[0]] = parts[1]
			}
		}
	} else {
		// Default selector
		selectorMap["app"] = serviceName
	}

	// Create service object
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": serviceName,
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceType(serviceType),
			Selector: selectorMap,
			Ports: []corev1.ServicePort{
				{
					Port:       port,
					TargetPort: intstr.FromInt(int(targetPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	// Create the service
	_, err = client.GetClientset().CoreV1().Services(namespace).Create(
		context.TODO(),
		service,
		metav1.CreateOptions{},
	)
	if err != nil {
		return fmt.Errorf("error creating service: %w", err)
	}

	fmt.Printf("✅ Service '%s' created successfully in namespace '%s'\n", serviceName, namespace)
	fmt.Printf("   Type: %s\n", serviceType)
	fmt.Printf("   Port: %d -> %d\n", port, targetPort)
	fmt.Printf("   Selector: %v\n", selectorMap)

	return nil
}

// Helper function to split key=value pairs
func splitKeyValue(pair string) []string {
	for i, char := range pair {
		if char == '=' {
			return []string{pair[:i], pair[i+1:]}
		}
	}
	return []string{pair}
}
