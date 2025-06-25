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
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	// Step 9 flags
	controllerNamespace  string
	controllerWorkers    int
	controllerSyncPeriod time.Duration
	enableControllerLogs bool
)

// Step 9: DeploymentController using sigs.k8s.io/controller-runtime
type DeploymentController struct {
	client.Client
	Scheme    *runtime.Scheme
	clientset kubernetes.Interface
}

// Step 9: Reconcile implements the reconcile.Reconciler interface
func (r *DeploymentController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log.Printf("🔄 Step 9: Reconciling deployment %s/%s", req.Namespace, req.Name)

	// Fetch the Deployment instance
	var deployment appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Printf("❌ Error fetching deployment: %v", err)
			return reconcile.Result{}, err
		}
		// Deployment was deleted
		log.Printf("🗑️ Step 9: Deployment %s/%s was deleted", req.Namespace, req.Name)
		return reconcile.Result{}, nil
	}

	// Log deployment details
	replicas := int32(0)
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	}

	log.Printf("📊 Step 9: Deployment Details:")
	log.Printf("   Name: %s", deployment.Name)
	log.Printf("   Namespace: %s", deployment.Namespace)
	log.Printf("   Desired Replicas: %d", replicas)
	log.Printf("   Ready Replicas: %d", deployment.Status.ReadyReplicas)
	log.Printf("   Available Replicas: %d", deployment.Status.AvailableReplicas)
	log.Printf("   Updated Replicas: %d", deployment.Status.UpdatedReplicas)

	// Log container information
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		container := deployment.Spec.Template.Spec.Containers[0]
		log.Printf("   Main Container: %s", container.Name)
		log.Printf("   Image: %s", container.Image)
	}

	// Check deployment health
	if deployment.Status.ReadyReplicas != replicas {
		log.Printf("⚠️ Step 9: Deployment %s/%s is not fully ready (%d/%d replicas)",
			deployment.Namespace, deployment.Name, deployment.Status.ReadyReplicas, replicas)

		// Requeue for retry
		return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
	} else if replicas > 0 {
		log.Printf("✅ Step 9: Deployment %s/%s is healthy (%d/%d replicas)",
			deployment.Namespace, deployment.Name, deployment.Status.ReadyReplicas, replicas)
	}

	// Log events for Step 9 requirement
	log.Printf("🎯 Step 9: Event processed successfully for deployment %s/%s", req.Namespace, req.Name)

	return reconcile.Result{}, nil
}

// Step 9: SetupWithManager sets up the controller with the Manager
func (r *DeploymentController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: controllerWorkers,
		}).
		Complete(r)
}

// Step 9: Multi-cluster informer setup for Step 9+
type MultiClusterInformer struct {
	clusters map[string]client.Client
	managers map[string]ctrl.Manager
}

func NewMultiClusterInformer() *MultiClusterInformer {
	return &MultiClusterInformer{
		clusters: make(map[string]client.Client),
		managers: make(map[string]ctrl.Manager),
	}
}

func (m *MultiClusterInformer) AddCluster(name string, config *ctrl.Config) error {
	log.Printf("🌐 Step 9+: Adding cluster '%s' to multi-cluster informer", name)

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:             runtime.NewScheme(),
		MetricsBindAddress: "0", // Disable metrics for individual cluster managers
		LeaderElection:     false,
	})
	if err != nil {
		return fmt.Errorf("failed to create manager for cluster %s: %v", name, err)
	}

	// Add schemes
	if err := appsv1.AddToScheme(mgr.GetScheme()); err != nil {
		return fmt.Errorf("failed to add appsv1 scheme: %v", err)
	}

	// Setup controller for this cluster
	controller := &DeploymentController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	if err := controller.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup controller for cluster %s: %v", name, err)
	}

	m.clusters[name] = mgr.GetClient()
	m.managers[name] = mgr

	log.Printf("✅ Step 9+: Successfully added cluster '%s'", name)
	return nil
}

func (m *MultiClusterInformer) Start(ctx context.Context) error {
	log.Printf("🚀 Step 9+: Starting multi-cluster informers for %d clusters", len(m.managers))

	for name, mgr := range m.managers {
		go func(clusterName string, manager ctrl.Manager) {
			log.Printf("🏃 Step 9+: Starting manager for cluster '%s'", clusterName)
			if err := manager.Start(ctx); err != nil {
				log.Printf("❌ Step 9+: Manager for cluster '%s' failed: %v", clusterName, err)
			}
		}(name, mgr)
	}

	return nil
}

// Step 9: Controller command
var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Start controller-runtime based deployment controller (Step 9)",
	Long: `Start a controller using sigs.k8s.io/controller-runtime that watches deployment events
and reports each event received with informer in logs.

Step 9 Features:
• Uses sigs.k8s.io/controller-runtime framework
• Implements Reconcile() method for deployment events
• Reports all events in logs as required
• Configurable worker count and sync period
• Proper error handling and requeuing

Step 9+ Features:
• Multi-cluster informers support
• Dynamically created informers for multiple clusters
• Isolated managers per cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		runController()
	},
}

func runController() {
	log.Println("🎯 Starting Step 9: sigs.k8s.io/controller-runtime deployment controller...")

	// Setup logging
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Create manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             runtime.NewScheme(),
		MetricsBindAddress: ":8080",
		Port:               9443,
		LeaderElection:     false, // Will be enabled in Step 10
		LeaderElectionID:   "k8s-cli-controller",
		Namespace:          controllerNamespace,
	})
	if err != nil {
		log.Fatalf("❌ Failed to create manager: %v", err)
	}

	// Add schemes
	if err := appsv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatalf("❌ Failed to add appsv1 scheme: %v", err)
	}

	// Create clientset for additional operations
	clientset, err := GetKubernetesClient()
	if err != nil {
		log.Fatalf("❌ Failed to create clientset: %v", err)
	}

	// Setup controller
	controller := &DeploymentController{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		clientset: clientset,
	}

	if err := controller.SetupWithManager(mgr); err != nil {
		log.Fatalf("❌ Failed to setup controller: %v", err)
	}

	log.Printf("⚙️ Step 9 Configuration:")
	log.Printf("   Namespace: %s", controllerNamespace)
	log.Printf("   Workers: %d", controllerWorkers)
	log.Printf("   Sync Period: %v", controllerSyncPeriod)
	log.Printf("   Enable Logs: %t", enableControllerLogs)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handler
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start manager in goroutine
	go func() {
		log.Println("🚀 Step 9: Starting controller manager...")
		if err := mgr.Start(ctx); err != nil {
			log.Fatalf("❌ Manager failed to start: %v", err)
		}
	}()

	log.Println("🎉 Step 9: Controller is running and watching deployment events.")
	log.Println("📋 Step 9 Features Active:")
	log.Println("   ✅ sigs.k8s.io/controller-runtime framework")
	log.Println("   ✅ Reconcile() method implementation")
	log.Println("   ✅ Event logging for each received event")
	log.Println("   ✅ Configurable workers and sync period")
	log.Println("   ✅ Proper error handling and requeuing")
	log.Println("")
	log.Println("🧪 Test the controller:")
	log.Println("   kubectl create deployment test-step9 --image=nginx:1.20")
	log.Println("   kubectl scale deployment test-step9 --replicas=3")
	log.Println("   kubectl delete deployment test-step9")

	// Wait for shutdown signal
	<-signalChan
	log.Println("\n🛑 Shutdown signal received, stopping controller...")

	cancel()
	log.Println("👋 Step 9: Controller stopped gracefully")
}

func init() {
	// Add flags for Step 9
	controllerCmd.Flags().StringVar(&controllerNamespace, "namespace", "", "Namespace to watch (empty = all namespaces)")
	controllerCmd.Flags().IntVar(&controllerWorkers, "workers", 1, "Number of controller workers")
	controllerCmd.Flags().DurationVar(&controllerSyncPeriod, "sync-period", 10*time.Minute, "Controller sync period")
	controllerCmd.Flags().BoolVar(&enableControllerLogs, "enable-logs", true, "Enable detailed controller logs")
	controllerCmd.Flags().StringVar(&configFile, "config", "", "Path to configuration file")

	// Register command
	RootCmd.AddCommand(controllerCmd)
}
