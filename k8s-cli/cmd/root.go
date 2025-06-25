package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

var (
	kubeconfig string
	namespace  string
	output     string

	// Step 7/7+/7++ добавленные переменные
	inCluster bool
)

// rootCmd представляет базовую команду при вызове без подкоманд
var rootCmd = &cobra.Command{
	Use:   "k8s-cli",
	Short: "Простой CLI для работы с Kubernetes",
	Long: `k8s-cli - это простой инструмент командной строки для работы с Kubernetes кластерами.

Возможности:
• Переключение между контекстами Kubernetes
• Просмотр списка различных ресурсов (pods, deployments, services)
• Создание ресурсов из YAML файлов
• Step 7: Мониторинг deployments через информеры (watch-informer)
• Step 7+: JSON API для доступа к кешу информеров (api-server)
• Step 7++: Управление конфигурацией (config)`,
	Version: "1.0.0",
}

// Execute добавляет все дочерние команды к корневой команде и устанавливает флаги
func Execute() error {
	return rootCmd.Execute()
}

// GetKubernetesClient - экспортируемая функция для получения клиента
// Используется в Step 7/7+/7++ функционале
func GetKubernetesClient() (kubernetes.Interface, error) {
	var config *rest.Config
	var err error

	if inCluster {
		config, err = rest.InClusterConfig()
	} else {
		// Используем существующий kubeconfig
		configPath := kubeconfig
		if configPath == "" {
			configPath = viper.GetString("kubeconfig")
		}
		config, err = clientcmd.BuildConfigFromFlags("", configPath)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create config: %v", err)
	}

	// Настройки производительности для Step 7+
	config.QPS = 50
	config.Burst = 100

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return clientset, nil
}

// RootCmd экспортируем для использования в других файлах
var RootCmd = rootCmd

func init() {
	cobra.OnInitialize(initConfig)

	// Существующие глобальные флаги
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "путь к kubeconfig файлу")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "namespace для операций")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "формат вывода (table, json, yaml)")

	// Step 7: Добавляем флаг для in-cluster режима
	rootCmd.PersistentFlags().BoolVar(&inCluster, "in-cluster", false, "использовать in-cluster аутентификацию")

	// Привязать флаги к viper
	viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("in-cluster", rootCmd.PersistentFlags().Lookup("in-cluster"))
}

func initConfig() {
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

	// Step 7++: Добавляем дополнительные пути для конфигурации
	viper.AddConfigPath(filepath.Join(homedir.HomeDir(), ".k8s-cli"))
	viper.AddConfigPath("/etc/k8s-cli")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Используется конфигурационный файл:", viper.ConfigFileUsed())
	}
}
