package cmd

import (
	"fmt"
	"io/ioutil"
	"k8s-cli/internal/k8s"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applyCmd представляет команду apply
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Создать ресурсы из YAML файла",
	Long:  "Создать или обновить ресурсы Kubernetes из YAML файла",
}

// applyFileCmd создает ресурсы из файла
var applyFileCmd = &cobra.Command{
	Use:   "file <filename>",
	Short: "Применить YAML файл",
	Long:  "Создать или обновить ресурсы Kubernetes из указанного YAML файла",
	Args:  cobra.ExactArgs(1),
	Example: `  # Применить YAML файл
  k8s-cli apply file pod.yaml

  # Применить файл в определенном namespace
  k8s-cli apply file deployment.yaml -n my-app`,
	RunE: runApplyFile,
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.AddCommand(applyFileCmd)
}

func runApplyFile(cmd *cobra.Command, args []string) error {
	filename := args[0]

	// Читаем YAML файл
	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла %s: %w", filename, err)
	}

	// Создаем Kubernetes клиент
	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("ошибка создания клиента: %w", err)
	}

	namespace := viper.GetString("namespace")

	// Применяем YAML
	err = client.CreateFromYAML(yamlData, namespace)
	if err != nil {
		return fmt.Errorf("ошибка применения YAML: %w", err)
	}

	fmt.Printf("✅ Ресурсы успешно созданы из файла: %s\n", filename)
	return nil
}
