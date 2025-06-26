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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	// Step 10 flags
	enableLeaderElection bool
	leaderElectionID     string
	managerMetricsPort   int
	managerHealthPort    int
	managerNamespace     string
)

// Step 10: Enhanced manager configuration
type ManagerConfig struct {
	LeaderElection   bool
	LeaderElectionID string
	MetricsPort      int
	HealthPort       int
	Namespace        string
	Workers          int
}

// Step 10: Controller Manager
type ControllerManager struct {
	manager ctrl.Manager
	config  *ManagerConfig
}

func NewControllerManager(config *ManagerConfig) (*ControllerManager, error) {
	log.Printf("🏗️ Step 10: Creating controller manager with configuration:")
	log.Printf("   Leader Election: %t", config.LeaderElection)
	log.Printf("   Leader Election ID: %s", config.LeaderElectionID)
	log.Printf("   Metrics Port: %d", config.MetricsPort)
	log.Printf("   Health Port: %d", config.HealthPort)
	log.Printf("   Namespace: %s", config.Namespace)
	log.Printf("   Workers: %d", config.Workers)

	// Setup manager options
	options := ctrl.Options{
		Scheme: runtime.NewScheme(),
		Metrics: server.Options{
			BindAddress: fmt.Sprintf(":%d", config.MetricsPort),
		},
		HealthProbeBindAddress:  fmt.Sprintf(":%d", config.HealthPort),
		LeaderElection:          config.LeaderElection,
		LeaderElectionID:        config.LeaderElectionID,
		LeaderElectionNamespace: config.Namespace,
	}

	// Set namespace if specified
	if config.Namespace != "" {
		options.Cache = cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				config.Namespace: {},
			},
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %v", err)
	}

	// Add schemes
	if err := appsv1.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, fmt.Errorf("failed to add appsv1 scheme: %v", err)
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("failed to add health check: %v", err)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("failed to add ready check: %v", err)
	}

	return &ControllerManager{
		manager: mgr,
		config:  config,
	}, nil
}

func (cm *ControllerManager) SetupControllers() error {
	log.Println("🔧 Step 10: Setting up controllers...")

	// Setup Deployment Controller
	deploymentController := &DeploymentController{
		Client: cm.manager.GetClient(),
		Scheme: cm.manager.GetScheme(),
	}

	if err := deploymentController.SetupWithManager(cm.manager); err != nil {
		return fmt.Errorf("failed to setup deployment controller: %v", err)
	}

	log.Println("✅ Step 10: Deployment controller registered")
	return nil
}

func (cm *ControllerManager) Start(ctx context.Context) error {
	log.Println("🚀 Step 10: Starting controller manager...")

	if cm.config.LeaderElection {
		log.Printf("🗳️ Step 10: Leader election enabled with ID: %s", cm.config.LeaderElectionID)
		log.Println("   📋 Manager will compete for leadership using lease resource")
		log.Println("   📋 Only the leader will process events")
	} else {
		log.Println("⚠️ Step 10: Leader election is disabled")
		log.Println("   📋 Manager will start immediately without election")
	}

	return cm.manager.Start(ctx)
}

func (cm *ControllerManager) GetManager() ctrl.Manager {
	return cm.manager
}

// Step 10: Manager command
var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Start controller manager with leader election (Step 10)",
	Long: `Start a controller manager that controls informers and controllers with leader election support.

Step 10 Features:
• Controller manager to control informer and controller
• Leader election with lease resource
• Flag to disable leader election
• Flag for manager metrics port
• Health checks and readiness probes
• Multi-controller management
• Graceful shutdown handling

Leader Election:
• Uses Kubernetes lease resource for coordination
• Only one manager instance processes events at a time
• Automatic failover when leader goes down
• Configurable lease duration and renew deadline`,
	Run: func(cmd *cobra.Command, args []string) {
		runManager()
	},
}

func runManager() {
	log.Println("🎯 Starting Step 10: Controller Manager with Leader Election...")

	// Setup logging
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Create manager configuration
	config := &ManagerConfig{
		LeaderElection:   enableLeaderElection,
		LeaderElectionID: leaderElectionID,
		MetricsPort:      managerMetricsPort,
		HealthPort:       managerHealthPort,
		Namespace:        managerNamespace,
		Workers:          controllerWorkers,
	}

	// Create controller manager
	cm, err := NewControllerManager(config)
	if err != nil {
		log.Fatalf("❌ Failed to create controller manager: %v", err)
	}

	// Setup controllers
	if err := cm.SetupControllers(); err != nil {
		log.Fatalf("❌ Failed to setup controllers: %v", err)
	}

	// Setup context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start manager in goroutine
	go func() {
		if err := cm.Start(ctx); err != nil {
			log.Fatalf("❌ Manager failed to start: %v", err)
		}
	}()

	log.Println("🎉 Step 10: Controller Manager is running!")
	log.Println("")
	log.Println("📋 Step 10 Features Active:")
	log.Println("   ✅ Controller manager controlling informers and controllers")
	if enableLeaderElection {
		log.Println("   ✅ Leader election with lease resource")
		log.Printf("   ✅ Leader election ID: %s", leaderElectionID)
	} else {
		log.Println("   ⚠️ Leader election disabled")
	}
	log.Printf("   ✅ Metrics server on port %d", managerMetricsPort)
	log.Printf("   ✅ Health checks on port %d", managerHealthPort)
	log.Println("   ✅ Graceful shutdown handling")
	log.Println("")
	log.Println("🔗 Endpoints:")
	log.Printf("   📊 Metrics: http://localhost:%d/metrics", managerMetricsPort)
	log.Printf("   ❤️ Health: http://localhost:%d/healthz", managerHealthPort)
	log.Printf("   ✅ Ready: http://localhost:%d/readyz", managerHealthPort)
	log.Println("")
	log.Println("🧪 Test the manager:")
	log.Println("   kubectl create deployment test-step10 --image=nginx:1.20")
	log.Println("   kubectl get leases -n kube-system | grep k8s-cli")
	log.Printf("   curl http://localhost:%d/healthz", managerHealthPort)
	log.Printf("   curl http://localhost:%d/metrics", managerMetricsPort)

	// Wait for shutdown signal
	<-signalChan
	log.Println("\n🛑 Shutdown signal received, stopping manager...")

	cancel()

	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)
	log.Println("👋 Step 10: Controller Manager stopped gracefully")
}

func init() {
	// Add flags for Step 10
	managerCmd.Flags().BoolVar(&enableLeaderElection, "enable-leader-election", true, "Enable leader election for manager")
	managerCmd.Flags().StringVar(&leaderElectionID, "leader-election-id", "k8s-cli-manager", "Leader election ID")
	managerCmd.Flags().IntVar(&managerMetricsPort, "metrics-port", 8080, "Port for metrics server")
	managerCmd.Flags().IntVar(&managerHealthPort, "health-port", 8081, "Port for health checks")
	managerCmd.Flags().StringVar(&managerNamespace, "manager-namespace", "", "Namespace for manager operations")
	managerCmd.Flags().IntVar(&controllerWorkers, "workers", 2, "Number of controller workers")

	// Register command
	RootCmd.AddCommand(managerCmd)
}
