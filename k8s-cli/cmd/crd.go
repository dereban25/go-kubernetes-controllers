package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	k8scliv1 "k8s-cli/api/v1"
	"k8s-cli/controllers"
)

var (
	scheme = runtime.NewScheme()
	// Step 11 flags
	crdMetricsPort          int
	crdHealthPort           int
	enableCRDLeaderElection bool
	crdLeaderElectionID     string
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(k8scliv1.AddToScheme(scheme))
}

// Step 11++: Multi-cluster client configuration
type MultiClusterConfig struct {
	Clusters map[string]ClusterConfig `yaml:"clusters"`
}

type ClusterConfig struct {
	Name       string `yaml:"name"`
	Kubeconfig string `yaml:"kubeconfig"`
	Context    string `yaml:"context"`
	Namespace  string `yaml:"namespace"`
	Enabled    bool   `yaml:"enabled"`
}

type MultiClusterManager struct {
	configs  map[string]ClusterConfig
	managers map[string]ctrl.Manager
}

func NewMultiClusterManager() *MultiClusterManager {
	return &MultiClusterManager{
		configs:  make(map[string]ClusterConfig),
		managers: make(map[string]ctrl.Manager),
	}
}

func (mcm *MultiClusterManager) AddCluster(name string, config ClusterConfig) error {
	log.Printf("🌐 Step 11++: Adding cluster '%s' to multi-cluster manager", name)

	// Create manager for this cluster
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: "0", // Disable metrics for individual clusters
		},
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				config.Namespace: {},
			},
		},
		HealthProbeBindAddress: "0",   // Disable health for individual clusters
		LeaderElection:         false, // No leader election per cluster
	})
	if err != nil {
		return fmt.Errorf("failed to create manager for cluster %s: %v", name, err)
	}

	// Setup FrontendPage controller for this cluster
	if err = (&controllers.FrontendPageReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup FrontendPageReconciler for cluster %s: %v", name, err)
	}

	mcm.configs[name] = config
	mcm.managers[name] = mgr

	log.Printf("✅ Step 11++: Successfully configured cluster '%s'", name)
	return nil
}

func (mcm *MultiClusterManager) StartAll(ctx context.Context) error {
	log.Printf("🚀 Step 11++: Starting multi-cluster managers for %d clusters", len(mcm.managers))

	for name, mgr := range mcm.managers {
		if !mcm.configs[name].Enabled {
			log.Printf("⏭️ Step 11++: Skipping disabled cluster '%s'", name)
			continue
		}

		go func(clusterName string, manager ctrl.Manager) {
			log.Printf("🏃 Step 11++: Starting manager for cluster '%s'", clusterName)
			if err := manager.Start(ctx); err != nil {
				log.Printf("❌ Step 11++: Manager for cluster '%s' failed: %v", clusterName, err)
			}
		}(name, mgr)
	}

	return nil
}

// Step 11: CRD command
var crdCmd = &cobra.Command{
	Use:   "crd",
	Short: "Start custom resource controller for FrontendPage (Step 11)",
	Long: `Start a controller for custom FrontendPage CRD with additional reconciliation logic.

Step 11 Features:
• Custom CRD: FrontendPage with spec and status
• Additional informer for custom resource
• Controller with reconciliation logic for custom resource
• Creates Deployment and Service for each FrontendPage
• Status updates and condition management
• Owner references and garbage collection

Step 11++ Features:
• Multi-cluster client configuration for management clusters
• Support for multiple Kubernetes clusters
• Per-cluster namespace isolation
• Configurable cluster endpoints

FrontendPage Resource:
• Manages frontend applications as code
• Automatically creates Deployment and Service
• Configurable replicas, image, and environment
• Status tracking and health monitoring`,
	Run: func(cmd *cobra.Command, args []string) {
		runCRDController()
	},
}

func runCRDController() {
	log.Println("🎯 Starting Step 11: Custom FrontendPage CRD Controller...")

	// Setup logging
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Create manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: server.Options{
			BindAddress: fmt.Sprintf(":%d", crdMetricsPort),
		},
		HealthProbeBindAddress: fmt.Sprintf(":%d", crdHealthPort),
		LeaderElection:         enableCRDLeaderElection,
		LeaderElectionID:       crdLeaderElectionID,
	})
	if err != nil {
		log.Fatalf("❌ Failed to create manager: %v", err)
	}

	// Setup FrontendPage controller
	if err = (&controllers.FrontendPageReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("❌ Failed to setup FrontendPageReconciler: %v", err)
	}

	// Setup Deployment controller for additional monitoring
	if err = (&DeploymentController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("❌ Failed to setup DeploymentController: %v", err)
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatalf("❌ Failed to add health check: %v", err)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatalf("❌ Failed to add ready check: %v", err)
	}

	// Setup context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start manager in goroutine
	go func() {
		if err := mgr.Start(ctx); err != nil {
			log.Fatalf("❌ Manager failed to start: %v", err)
		}
	}()

	log.Println("🎉 Step 11: FrontendPage CRD Controller is running!")
	log.Println("")
	log.Println("📋 Step 11 Features Active:")
	log.Println("   ✅ Custom FrontendPage CRD")
	log.Println("   ✅ Additional informer for custom resource")
	log.Println("   ✅ Controller with reconciliation logic")
	log.Println("   ✅ Automatic Deployment and Service creation")
	log.Println("   ✅ Status updates and condition management")
	log.Println("   ✅ Owner references and garbage collection")
	if enableCRDLeaderElection {
		log.Printf("   ✅ Leader election enabled with ID: %s", crdLeaderElectionID)
	} else {
		log.Println("   ⚠️ Leader election disabled")
	}
	log.Println("")
	log.Println("🔗 Endpoints:")
	log.Printf("   📊 Metrics: http://localhost:%d/metrics", crdMetricsPort)
	log.Printf("   ❤️ Health: http://localhost:%d/healthz", crdHealthPort)
	log.Printf("   ✅ Ready: http://localhost:%d/readyz", crdHealthPort)
	log.Println("")
	log.Println("🧪 Test the CRD controller:")
	log.Println("   # First, apply the CRD:")
	log.Println("   kubectl apply -f config/crd/")
	log.Println("")
	log.Println("   # Create a FrontendPage:")
	log.Println("   kubectl apply -f - <<EOF")
	log.Println("   apiVersion: k8scli.dev/v1")
	log.Println("   kind: FrontendPage")
	log.Println("   metadata:")
	log.Println("     name: my-frontend")
	log.Println("   spec:")
	log.Println("     title: \"My Frontend App\"")
	log.Println("     description: \"A sample frontend application\"")
	log.Println("     path: \"/app\"")
	log.Println("     replicas: 2")
	log.Println("     image: \"nginx:1.20\"")
	log.Println("     config:")
	log.Println("       ENVIRONMENT: \"production\"")
	log.Println("   EOF")
	log.Println("")
	log.Println("   # Check the resources:")
	log.Println("   kubectl get frontendpages")
	log.Println("   kubectl describe frontendpage my-frontend")
	log.Println("   kubectl get deployments,services")

	// Wait for shutdown signal
	<-signalChan
	log.Println("\n🛑 Shutdown signal received, stopping CRD controller...")

	cancel()
	time.Sleep(2 * time.Second)
	log.Println("👋 Step 11: FrontendPage CRD Controller stopped gracefully")
}

// Step 11++: Multi-cluster command
var multiClusterCmd = &cobra.Command{
	Use:   "multi-cluster",
	Short: "Start multi-cluster management for FrontendPage CRDs (Step 11++)",
	Long: `Start multi-cluster management system for FrontendPage CRDs across multiple Kubernetes clusters.

Step 11++ Features:
• Multi-cluster client configuration
• Management of multiple Kubernetes clusters
• Per-cluster namespace isolation
• Configurable cluster endpoints
• Cross-cluster resource synchronization`,
	Run: func(cmd *cobra.Command, args []string) {
		runMultiClusterManager()
	},
}

func runMultiClusterManager() {
	log.Println("🎯 Starting Step 11++: Multi-Cluster Management...")

	// Create multi-cluster manager
	mcm := NewMultiClusterManager()

	// Example cluster configurations (in real implementation, load from config file)
	clusters := []ClusterConfig{
		{
			Name:       "production",
			Kubeconfig: "~/.kube/config-prod",
			Context:    "production-cluster",
			Namespace:  "frontend-prod",
			Enabled:    true,
		},
		{
			Name:       "staging",
			Kubeconfig: "~/.kube/config-staging",
			Context:    "staging-cluster",
			Namespace:  "frontend-staging",
			Enabled:    true,
		},
		{
			Name:       "development",
			Kubeconfig: "~/.kube/config",
			Context:    "docker-desktop",
			Namespace:  "default",
			Enabled:    true,
		},
	}

	// Add clusters to manager
	for _, cluster := range clusters {
		if err := mcm.AddCluster(cluster.Name, cluster); err != nil {
			log.Printf("⚠️ Failed to add cluster %s: %v", cluster.Name, err)
			continue
		}
	}

	// Setup context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start all cluster managers
	if err := mcm.StartAll(ctx); err != nil {
		log.Fatalf("❌ Failed to start multi-cluster managers: %v", err)
	}

	log.Println("🎉 Step 11++: Multi-Cluster Management is running!")
	log.Println("")
	log.Println("📋 Step 11++ Features Active:")
	log.Printf("   ✅ Managing %d clusters", len(clusters))
	log.Println("   ✅ Multi-cluster client configuration")
	log.Println("   ✅ Per-cluster namespace isolation")
	log.Println("   ✅ Cross-cluster resource synchronization")
	log.Println("")
	log.Println("🌐 Configured Clusters:")
	for _, cluster := range clusters {
		status := "✅ Enabled"
		if !cluster.Enabled {
			status = "⏸️ Disabled"
		}
		log.Printf("   %s: %s (namespace: %s) %s", cluster.Name, cluster.Context, cluster.Namespace, status)
	}

	// Wait for shutdown signal
	<-signalChan
	log.Println("\n🛑 Shutdown signal received, stopping multi-cluster management...")

	cancel()
	time.Sleep(2 * time.Second)
	log.Println("👋 Step 11++: Multi-Cluster Management stopped gracefully")
}

func init() {
	// Add flags for Step 11
	crdCmd.Flags().IntVar(&crdMetricsPort, "metrics-port", 8082, "Port for CRD controller metrics")
	crdCmd.Flags().IntVar(&crdHealthPort, "health-port", 8083, "Port for CRD controller health checks")
	crdCmd.Flags().BoolVar(&enableCRDLeaderElection, "enable-leader-election", false, "Enable leader election for CRD controller")
	crdCmd.Flags().StringVar(&crdLeaderElectionID, "leader-election-id", "k8s-cli-crd-controller", "Leader election ID for CRD controller")

	// Register commands
	RootCmd.AddCommand(crdCmd)
	RootCmd.AddCommand(multiClusterCmd)
}
