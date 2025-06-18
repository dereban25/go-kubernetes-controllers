package cmd

import (
	"fmt"
	"k8s-cli/internal/k8s"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// contextCmd представляет команду context
var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Управление контекстами Kubernetes",
	Long:  "Команды для работы с контекстами Kubernetes - просмотр, переключение",
}

// contextListCmd выводит список контекстов
var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "Список всех контекстов",
	Long:  "Показать все доступные контексты Kubernetes",
	Example: `  # Показать все контексты
  k8s-cli context list`,
	RunE: runContextList,
}

// contextCurrentCmd показывает текущий контекст
var contextCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Показать текущий контекст",
	Long:  "Показать текущий активный контекст Kubernetes",
	Example: `  # Показать текущий контекст
  k8s-cli context current`,
	RunE: runContextCurrent,
}

// contextSetCmd переключает контекст
var contextSetCmd = &cobra.Command{
	Use:   "set <context-name>",
	Short: "Переключить контекст",
	Long:  "Переключиться на указанный контекст Kubernetes",
	Args:  cobra.ExactArgs(1),
	Example: `  # Переключиться на контекст
  k8s-cli context set my-cluster`,
	RunE: runContextSet,
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextCurrentCmd)
	contextCmd.AddCommand(contextSetCmd)
}

func runContextList(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("ошибка создания клиента: %w", err)
	}

	contexts, err := client.GetContexts()
	if err != nil {
		return fmt.Errorf("ошибка получения контекстов: %w", err)
	}

	currentContext, err := client.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("ошибка получения текущего контекста: %w", err)
	}

	fmt.Println("Доступные контексты:")
	for _, context := range contexts {
		if context == currentContext {
			fmt.Printf("* %s (текущий)\n", context)
		} else {
			fmt.Printf("  %s\n", context)
		}
	}

	return nil
}

func runContextCurrent(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("ошибка создания клиента: %w", err)
	}

	currentContext, err := client.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("ошибка получения текущего контекста: %w", err)
	}

	fmt.Printf("Текущий контекст: %s\n", currentContext)
	return nil
}

func runContextSet(cmd *cobra.Command, args []string) error {
	contextName := args[0]
	kubeconfigPath := viper.GetString("kubeconfig")

	client, err := k8s.NewClient(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("ошибка создания клиента: %w", err)
	}

	err = client.SetContext(contextName, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("ошибка переключения контекста: %w", err)
	}

	fmt.Printf("Контекст переключен на: %s\n", contextName)
	return nil
}
