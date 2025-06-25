package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
)

// Step 7++: Enhanced configuration command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage k8s-cli configuration (Step 7++)",
	Long:  "Manage configuration files for k8s-cli informers and API server",
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	Long:  "Display the current configuration being used by k8s-cli",
	Run: func(cmd *cobra.Command, args []string) {
		viewConfig()
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  "Create a default configuration file for k8s-cli",
	Run: func(cmd *cobra.Command, args []string) {
		initConfigFile()
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  "Validate the syntax and values in configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		validateConfig()
	},
}

func viewConfig() {
	log.Println("📄 Current k8s-cli Configuration:")

	config, err := loadInformerConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	fmt.Printf(`
Step 7++ Configuration Status:
═══════════════════════════════════════════════════════════════

📋 Informer Settings:
   ResyncPeriod: %v
   Workers: %d
   Namespaces: %v
   LogEvents: %t

🌐 API Server:
   Enabled: %t
   Port: %d

🔧 Custom Logic:
   EnableUpdateHandling: %t
   EnableDeleteHandling: %t
   FilterLabels: %v

⚙️ Kubernetes Client:
   Timeout: %s
   QPS: %.0f
   Burst: %d

📝 Logging:
   Level: %s
   Format: %s

`,
		config.ResyncPeriod,
		config.Workers,
		config.Namespaces,
		config.LogEvents,
		config.APIServer.Enabled,
		config.APIServer.Port,
		config.CustomLogic.EnableUpdateHandling,
		config.CustomLogic.EnableDeleteHandling,
		config.CustomLogic.FilterLabels,
		config.Kubernetes.Timeout,
		config.Kubernetes.QPS,
		config.Kubernetes.Burst,
		config.Logging.Level,
		config.Logging.Format,
	)

	// Show which config file is being used
	if configFile != "" {
		fmt.Printf("📄 Config file: %s\n", configFile)
	} else {
		fmt.Println("📄 Using default configuration (no config file specified)")
		fmt.Println("💡 To create a config file, run: k8s-cli config init")
	}
}

func initConfigFile() {
	log.Println("🔧 Initializing k8s-cli configuration...")

	// Default configuration template
	configTemplate := `# k8s-cli Configuration File (Step 7++)
# This file configures the informer settings, API server, and custom logic

# Step 7: Informer settings
resync_period: "30s"  # How often to resync the cache
workers: 2            # Number of worker goroutines for processing events

# Namespaces to watch (empty means all namespaces)
namespaces:
  - "default"
  - "kube-system"

# Enable event logging
log_events: true

# Step 7+: JSON API Server configuration
api_server:
  enabled: true        # Enable JSON API server
  port: 8080          # API server port

# Step 7+: Custom logic configuration
custom_logic:
  # Enable custom handling for update events
  enable_update_handling: true
  
  # Enable custom handling for delete events  
  enable_delete_handling: true
  
  # Filter deployments by labels (optional)
  filter_labels:
    - "app"
    - "environment"

# Step 7++: Kubernetes client settings
kubernetes:
  # Timeout for API calls
  timeout: "30s"
  
  # QPS and burst settings for rate limiting
  qps: 50
  burst: 100

# Step 7++: Logging configuration
logging:
  level: "info"  # debug, info, warn, error
  format: "text" # text, json

# Example usage:
# k8s-cli watch-informer --config ~/.k8s-cli/config.yaml
# k8s-cli api-server --config ~/.k8s-cli/config.yaml
`

	// Determine config file path
	var configPath string
	if configFile != "" {
		configPath = configFile
	} else {
		// Use default location
		configDir := filepath.Join(homedir.HomeDir(), ".k8s-cli")
		configPath = filepath.Join(configDir, "config.yaml")

		// Create directory if it doesn't exist
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Fatalf("❌ Failed to create config directory: %v", err)
		}
	}

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		log.Printf("⚠️ Configuration file already exists at: %s", configPath)
		log.Println("💡 Use 'k8s-cli config view' to see current configuration")
		return
	}

	// Write configuration file
	if err := os.WriteFile(configPath, []byte(configTemplate), 0644); err != nil {
		log.Fatalf("❌ Failed to write config file: %v", err)
	}

	log.Printf("✅ Configuration file created at: %s", configPath)
	log.Println("📝 Edit the file to customize your settings")
	log.Println("🧪 Test your configuration with:")
	log.Printf("   k8s-cli config validate --config %s", configPath)
	log.Printf("   k8s-cli watch-informer --config %s", configPath)
}

func validateConfig() {
	log.Println("🔍 Validating k8s-cli configuration...")

	if configFile == "" {
		log.Println("⚠️ No config file specified. Using defaults.")
		log.Println("💡 Specify a config file with --config flag")
		return
	}

	// Check if file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatalf("❌ Config file not found: %s", configFile)
	}

	// Try to load and parse configuration
	config, err := loadInformerConfig()
	if err != nil {
		log.Fatalf("❌ Configuration validation failed: %v", err)
	}

	log.Printf("✅ Configuration file is valid: %s", configFile)

	// Validate specific settings
	fmt.Println("\n🔍 Configuration Validation Results:")
	fmt.Println("══════════════════════════════════════")

	// Validate resync period
	if config.ResyncPeriod < 1 {
		fmt.Println("❌ resync_period must be positive")
	} else {
		fmt.Printf("✅ resync_period: %v\n", config.ResyncPeriod)
	}

	// Validate workers
	if config.Workers < 1 || config.Workers > 10 {
		fmt.Println("⚠️ workers should be between 1 and 10")
	} else {
		fmt.Printf("✅ workers: %d\n", config.Workers)
	}

	// Validate API port
	if config.APIServer.Port < 1024 || config.APIServer.Port > 65535 {
		fmt.Println("⚠️ api_server.port should be between 1024 and 65535")
	} else {
		fmt.Printf("✅ api_server.port: %d\n", config.APIServer.Port)
	}

	// Validate namespaces
	if len(config.Namespaces) == 0 {
		fmt.Println("⚠️ No namespaces specified - will watch all namespaces")
	} else {
		fmt.Printf("✅ namespaces: %v\n", config.Namespaces)
	}

	fmt.Println("\n🎯 Step 7++ Features Enabled:")
	fmt.Printf("   📊 Custom update handling: %t\n", config.CustomLogic.EnableUpdateHandling)
	fmt.Printf("   🗑️ Custom delete handling: %t\n", config.CustomLogic.EnableDeleteHandling)
	fmt.Printf("   📝 Event logging: %t\n", config.LogEvents)
	fmt.Printf("   🌐 API server: %t\n", config.APIServer.Enabled)

	fmt.Println("\n✅ Configuration is ready to use!")
	fmt.Println("🚀 Test your configuration:")
	fmt.Printf("   k8s-cli watch-informer --config %s\n", configFile)
	fmt.Printf("   k8s-cli api-server --config %s\n", configFile)
}

// Step 7++: Find configuration file in standard locations
func findConfigFile() string {
	possibleConfigs := []string{
		configFile, // Command line specified
		filepath.Join(homedir.HomeDir(), ".k8s-cli", "config.yaml"),
		filepath.Join(homedir.HomeDir(), ".k8s-cli", "k8s-cli-config.yaml"),
		"/etc/k8s-cli/config.yaml",
		"/etc/k8s-cli/k8s-cli-config.yaml",
		"k8s-cli-config.yaml",
		"config.yaml",
	}

	for _, path := range possibleConfigs {
		if path != "" {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}

// Enhanced configuration loading with better path resolution for Step 7++
func loadInformerConfigEnhanced() (*InformerConfig, error) {
	config := &InformerConfig{
		ResyncPeriod: 30 * time.Second,
		Workers:      2,
		Namespaces:   []string{"default"},
		LogEvents:    true,
	}

	// Set defaults for all nested structs
	config.CustomLogic.EnableUpdateHandling = true
	config.CustomLogic.EnableDeleteHandling = true
	config.Kubernetes.Timeout = "30s"
	config.Kubernetes.QPS = 50
	config.Kubernetes.Burst = 100
	config.Logging.Level = "info"
	config.Logging.Format = "text"
	config.APIServer.Enabled = false
	config.APIServer.Port = 8080

	// Find config file
	foundConfigFile := findConfigFile()

	if foundConfigFile != "" {
		viper.SetConfigFile(foundConfigFile)
		if err := viper.ReadInConfig(); err == nil {
			log.Printf("📄 Using config file: %s", viper.ConfigFileUsed())
			if err := viper.Unmarshal(config); err != nil {
				return nil, fmt.Errorf("error unmarshaling config: %v", err)
			}
		} else {
			return nil, fmt.Errorf("error reading config file %s: %v", foundConfigFile, err)
		}
	}

	// Override with command line flags
	if informerResyncPeriod > 0 {
		config.ResyncPeriod = informerResyncPeriod
	}
	if informerWorkers > 0 {
		config.Workers = informerWorkers
	}
	if enableEventLogging {
		config.LogEvents = enableEventLogging
	}
	if apiPort > 0 {
		config.APIServer.Port = apiPort
	}

	return config, nil
}

func init() {
	// Add subcommands
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)

	// Add flags
	configCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to configuration file")

	// Register main command
	RootCmd.AddCommand(configCmd)
}
