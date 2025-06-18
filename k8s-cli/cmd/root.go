package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

var (
	kubeconfig string
	namespace  string
	output     string
)

// rootCmd представляет базовую команду при вызове без подкоманд
var rootCmd = &cobra.Command{
	Use:   "k8s-cli",
	Short: "Простой CLI для работы с Kubernetes",
	Long: `k8s-cli - это простой инструмент командной строки для работы с Kubernetes кластерами.

Возможности:
• Переключение между контекстами Kubernetes
• Просмотр списка различных ресурсов (pods, deployments, services)
• Создание ресурсов из YAML файлов`,
	Version: "1.0.0",
}

// Execute добавляет все дочерние команды к корневой команде и устанавливает флаги
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Глобальные флаги
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "путь к kubeconfig файлу")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "namespace для операций")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "формат вывода (table, json, yaml)")

	// Привязать флаги к viper
	viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}

func initConfig() {
	// Установить путь к kubeconfig по умолчанию
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
			viper.Set("kubeconfig", kubeconfig)
		}
	}

	// Настроить переменные окружения
	viper.SetEnvPrefix("K8S_CLI")
	viper.AutomaticEnv()

	// Прочитать конфигурационный файл если он существует
	viper.SetConfigName(".k8s-cli")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Используется конфигурационный файл:", viper.ConfigFileUsed())
	}
}
