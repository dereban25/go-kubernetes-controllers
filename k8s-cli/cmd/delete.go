package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"github.com/dereban25/go-kubernetes-controllers/k8s-cli/internal/k8s"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete Kubernetes resources",
	Long:  "Delete Kubernetes resources from the cluster",
}

// deleteFileCmd deletes resources from a YAML file
var deleteFileCmd = &cobra.Command{
	Use:   "file <filename>",
	Short: "Delete resources from YAML file",
	Long:  "Delete Kubernetes resources specified in a YAML file",
	Args:  cobra.ExactArgs(1),
	Example: `  # Delete resources from YAML file
  k8s-cli delete file pod.yaml

  # Delete resources from file in specific namespace
  k8s-cli delete file deployment.yaml -n my-app

  # Force delete without confirmation
  k8s-cli delete file pod.yaml --force`,
	RunE: runDeleteFile,
}

// deletePodCmd deletes a specific pod
var deletePodCmd = &cobra.Command{
	Use:   "pod <pod-name>",
	Short: "Delete a pod",
	Long:  "Delete a specific pod by name",
	Args:  cobra.ExactArgs(1),
	Example: `  # Delete a pod
  k8s-cli delete pod nginx-pod

  # Delete a pod in specific namespace
  k8s-cli delete pod nginx-pod -n my-app

  # Force delete without confirmation
  k8s-cli delete pod nginx-pod --force`,
	RunE: runDeletePod,
}

// deleteDeploymentCmd deletes a specific deployment
var deleteDeploymentCmd = &cobra.Command{
	Use:   "deployment <deployment-name>",
	Short: "Delete a deployment",
	Long:  "Delete a specific deployment by name",
	Args:  cobra.ExactArgs(1),
	Example: `  # Delete a deployment
  k8s-cli delete deployment nginx-deployment

  # Delete a deployment in specific namespace
  k8s-cli delete deployment nginx-deployment -n my-app

  # Force delete without confirmation
  k8s-cli delete deployment nginx-deployment --force`,
	RunE: runDeleteDeployment,
}

// deleteServiceCmd deletes a specific service
var deleteServiceCmd = &cobra.Command{
	Use:   "service <service-name>",
	Short: "Delete a service",
	Long:  "Delete a specific service by name",
	Args:  cobra.ExactArgs(1),
	Example: `  # Delete a service
  k8s-cli delete service nginx-service

  # Delete a service in specific namespace
  k8s-cli delete service nginx-service -n my-app

  # Force delete without confirmation
  k8s-cli delete service nginx-service --force`,
	RunE: runDeleteService,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteFileCmd)
	deleteCmd.AddCommand(deletePodCmd)
	deleteCmd.AddCommand(deleteDeploymentCmd)
	deleteCmd.AddCommand(deleteServiceCmd)

	// Add flags
	deleteFileCmd.Flags().Bool("force", false, "Force delete without confirmation")
	deletePodCmd.Flags().Bool("force", false, "Force delete without confirmation")
	deleteDeploymentCmd.Flags().Bool("force", false, "Force delete without confirmation")
	deleteServiceCmd.Flags().Bool("force", false, "Force delete without confirmation")
}

func runDeleteFile(cmd *cobra.Command, args []string) error {
	filename := args[0]
	force, _ := cmd.Flags().GetBool("force")

	// Read YAML file
	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	// Create Kubernetes client
	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	namespace := viper.GetString("namespace")

	// Parse YAML to get resource info
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(string(yamlData)), 4096)
	var obj unstructured.Unstructured
	if err := decoder.Decode(&obj); err != nil {
		return fmt.Errorf("error decoding YAML: %w", err)
	}

	// Confirm deletion unless force flag is used
	if !force {
		fmt.Printf("Are you sure you want to delete %s/%s? (y/N): ", obj.GetKind(), obj.GetName())
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Delete the resource
	err = client.DeleteFromYAML(yamlData, namespace)
	if err != nil {
		return fmt.Errorf("error deleting resource: %w", err)
	}

	fmt.Printf("✅ Resource successfully deleted from file: %s\n", filename)
	return nil
}

func runDeletePod(cmd *cobra.Command, args []string) error {
	podName := args[0]
	force, _ := cmd.Flags().GetBool("force")
	namespace := viper.GetString("namespace")

	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	// Confirm deletion unless force flag is used
	if !force {
		fmt.Printf("Are you sure you want to delete pod/%s in namespace %s? (y/N): ", podName, namespace)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Delete the pod
	err = client.GetClientset().CoreV1().Pods(namespace).Delete(
		context.TODO(),
		podName,
		metav1.DeleteOptions{},
	)
	if err != nil {
		return fmt.Errorf("error deleting pod: %w", err)
	}

	fmt.Printf("✅ Pod '%s' successfully deleted from namespace '%s'\n", podName, namespace)
	return nil
}

func runDeleteDeployment(cmd *cobra.Command, args []string) error {
	deploymentName := args[0]
	force, _ := cmd.Flags().GetBool("force")
	namespace := viper.GetString("namespace")

	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	// Confirm deletion unless force flag is used
	if !force {
		fmt.Printf("Are you sure you want to delete deployment/%s in namespace %s? (y/N): ", deploymentName, namespace)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Delete the deployment
	err = client.GetClientset().AppsV1().Deployments(namespace).Delete(
		context.TODO(),
		deploymentName,
		metav1.DeleteOptions{},
	)
	if err != nil {
		return fmt.Errorf("error deleting deployment: %w", err)
	}

	fmt.Printf("✅ Deployment '%s' successfully deleted from namespace '%s'\n", deploymentName, namespace)
	return nil
}

func runDeleteService(cmd *cobra.Command, args []string) error {
	serviceName := args[0]
	force, _ := cmd.Flags().GetBool("force")
	namespace := viper.GetString("namespace")

	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	// Confirm deletion unless force flag is used
	if !force {
		fmt.Printf("Are you sure you want to delete service/%s in namespace %s? (y/N): ", serviceName, namespace)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Delete the service
	err = client.GetClientset().CoreV1().Services(namespace).Delete(
		context.TODO(),
		serviceName,
		metav1.DeleteOptions{},
	)
	if err != nil {
		return fmt.Errorf("error deleting service: %w", err)
	}

	fmt.Printf("✅ Service '%s' successfully deleted from namespace '%s'\n", serviceName, namespace)
	return nil
}
