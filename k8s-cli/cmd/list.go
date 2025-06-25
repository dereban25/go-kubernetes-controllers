package cmd

import (
	"context"
	"fmt"
	"github.com/dereban25/go-kubernetes-controllers/k8s-cli/internal/k8s"
	"github.com/dereban25/go-kubernetes-controllers/k8s-cli/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// listCmd представляет команду list
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Список ресурсов Kubernetes",
	Long:  "Команды для получения списков различных ресурсов Kubernetes",
}

// listPodsCmd выводит список подов
var listPodsCmd = &cobra.Command{
	Use:   "pods",
	Short: "Список подов",
	Long:  "Показать все поды в указанном namespace",
	Example: `  # Список подов в текущем namespace
  k8s-cli list pods

  # Список подов в определенном namespace
  k8s-cli list pods -n kube-system

  # Вывод в JSON формате
  k8s-cli list pods -o json`,
	RunE: runListPods,
}

// listDeploymentsCmd выводит список деплойментов
var listDeploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "Список деплойментов",
	Long:  "Показать все деплойменты в указанном namespace",
	Example: `  # Список деплойментов
  k8s-cli list deployments

  # Список деплойментов в определенном namespace
  k8s-cli list deployments -n my-app`,
	RunE: runListDeployments,
}

// listServicesCmd выводит список сервисов
var listServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Список сервисов",
	Long:  "Показать все сервисы в указанном namespace",
	Example: `  # Список сервисов
  k8s-cli list services

  # Список сервисов в определенном namespace
  k8s-cli list services -n production`,
	RunE: runListServices,
}

// listNamespacesCmd выводит список namespace'ов
var listNamespacesCmd = &cobra.Command{
	Use:   "namespaces",
	Short: "Список namespace'ов",
	Long:  "Показать все namespace'ы в кластере",
	Example: `  # Список namespace'ов
  k8s-cli list namespaces`,
	RunE: runListNamespaces,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listPodsCmd)
	listCmd.AddCommand(listDeploymentsCmd)
	listCmd.AddCommand(listServicesCmd)
	listCmd.AddCommand(listNamespacesCmd)

	// Добавляем флаги
	listPodsCmd.Flags().StringP("selector", "l", "", "селектор меток")
	listDeploymentsCmd.Flags().StringP("selector", "l", "", "селектор меток")
	listServicesCmd.Flags().StringP("selector", "l", "", "селектор меток")
}

func runListPods(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("ошибка создания клиента: %w", err)
	}

	namespace := viper.GetString("namespace")
	selector, _ := cmd.Flags().GetString("selector")

	listOptions := metav1.ListOptions{}
	if selector != "" {
		listOptions.LabelSelector = selector
	}

	pods, err := client.GetClientset().CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return fmt.Errorf("ошибка получения подов: %w", err)
	}

	fmt.Printf("Поды в namespace '%s':\n", namespace)
	return utils.PrintPods(pods.Items, viper.GetString("output"))
}

func runListDeployments(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("ошибка создания клиента: %w", err)
	}

	namespace := viper.GetString("namespace")
	selector, _ := cmd.Flags().GetString("selector")

	listOptions := metav1.ListOptions{}
	if selector != "" {
		listOptions.LabelSelector = selector
	}

	deployments, err := client.GetClientset().AppsV1().Deployments(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return fmt.Errorf("ошибка получения деплойментов: %w", err)
	}

	fmt.Printf("Деплойменты в namespace '%s':\n", namespace)
	return utils.PrintDeployments(deployments.Items, viper.GetString("output"))
}

func runListServices(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("ошибка создания клиента: %w", err)
	}

	namespace := viper.GetString("namespace")
	selector, _ := cmd.Flags().GetString("selector")

	listOptions := metav1.ListOptions{}
	if selector != "" {
		listOptions.LabelSelector = selector
	}

	services, err := client.GetClientset().CoreV1().Services(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return fmt.Errorf("ошибка получения сервисов: %w", err)
	}

	fmt.Printf("Сервисы в namespace '%s':\n", namespace)
	return utils.PrintServices(services.Items, viper.GetString("output"))
}

func runListNamespaces(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("ошибка создания клиента: %w", err)
	}

	namespaces, err := client.GetClientset().CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("ошибка получения namespace'ов: %w", err)
	}

	fmt.Println("Namespace'ы:")
	for _, ns := range namespaces.Items {
		fmt.Printf("  %s\n", ns.Name)
	}

	return nil
}
