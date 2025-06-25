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
	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var (
	// Step 7 flags
	informerResyncPeriod time.Duration
	informerWorkers      int
	enableEventLogging   bool
	configFile           string
)

// Step 7: Informer configuration structure
type InformerConfig struct {
	ResyncPeriod time.Duration `mapstructure:"resync_period"`
	Workers      int           `mapstructure:"workers"`
	Namespaces   []string      `mapstructure:"namespaces"`
	LogEvents    bool          `mapstructure:"log_events"`

	APIServer struct {
		Enabled bool `mapstructure:"enabled"`
		Port    int  `mapstructure:"port"`
	} `mapstructure:"api_server"`

	CustomLogic struct {
		EnableUpdateHandling bool     `mapstructure:"enable_update_handling"`
		EnableDeleteHandling bool     `mapstructure:"enable_delete_handling"`
		FilterLabels         []string `mapstructure:"filter_labels"`
	} `mapstructure:"custom_logic"`

	// Step 7++: Additional configuration
	Kubernetes struct {
		Timeout string  `mapstructure:"timeout"`
		QPS     float32 `mapstructure:"qps"`
		Burst   int     `mapstructure:"burst"`
	} `mapstructure:"kubernetes"`

	Logging struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"logging"`
}

// Step 7: Event processor for informers using k8s.io/client-go
type EventProcessor struct {
	clientset       kubernetes.Interface
	workqueue       workqueue.RateLimitingInterface
	config          *InformerConfig
	informerStop    chan struct{}
	deploymentCache map[string]*appsv1.Deployment
	cacheIndexer    cache.Indexer
	startTime       time.Time
}

func NewEventProcessor(clientset kubernetes.Interface, config *InformerConfig) *EventProcessor {
	return &EventProcessor{
		clientset:       clientset,
		workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "deployments"),
		config:          config,
		informerStop:    make(chan struct{}),
		deploymentCache: make(map[string]*appsv1.Deployment),
		startTime:       time.Now(),
	}
}

// Step 7: Start informer using k8s.io/client-go informers
func (e *EventProcessor) Start(ctx context.Context) error {
	log.Println("🚀 Starting Kubernetes deployment informer with k8s.io/client-go...")

	// Step 7: Create SharedInformerFactory for list/watch informer
	informerFactory := informers.NewSharedInformerFactory(e.clientset, e.config.ResyncPeriod)
	deploymentInformer := informerFactory.Apps().V1().Deployments().Informer()

	// Store indexer for direct cache access
	e.cacheIndexer = deploymentInformer.GetIndexer()

	// Step 7: Add event handlers for informer
	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if deployment, ok := obj.(*appsv1.Deployment); ok {
				e.handleAddEvent(deployment)
				// Step 7: Report events in logs
				if e.config.LogEvents {
					log.Printf("✅ ADD: Deployment %s/%s created", deployment.Namespace, deployment.Name)
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if oldDeployment, ok := oldObj.(*appsv1.Deployment); ok {
				if newDeployment, ok := newObj.(*appsv1.Deployment); ok {
					e.handleUpdateEvent(oldDeployment, newDeployment)
					// Step 7: Report events in logs
					if e.config.LogEvents {
						log.Printf("🔄 UPDATE: Deployment %s/%s modified", newDeployment.Namespace, newDeployment.Name)
					}
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if deployment, ok := obj.(*appsv1.Deployment); ok {
				e.handleDeleteEvent(deployment)
				// Step 7: Report events in logs
				if e.config.LogEvents {
					log.Printf("🗑️ DELETE: Deployment %s/%s removed", deployment.Namespace, deployment.Name)
				}
			}
		},
	})

	// Step 7: Start informer factory
	informerFactory.Start(e.informerStop)

	log.Println("⏳ Waiting for informer cache to sync...")
	if !cache.WaitForCacheSync(ctx.Done(), deploymentInformer.HasSynced) {
		return fmt.Errorf("failed to sync informer cache")
	}
	log.Println("✅ Informer cache synced successfully")

	// Start worker goroutines
	for i := 0; i < e.config.Workers; i++ {
		go e.runWorker(ctx)
	}

	log.Printf("🔄 Started %d workers, watching deployment events...", e.config.Workers)
	return nil
}

func (e *EventProcessor) Stop() {
	log.Println("🛑 Stopping deployment informer...")
	close(e.informerStop)
	e.workqueue.ShutDown()
}

// Step 7+: Custom logic for handling events
func (e *EventProcessor) handleAddEvent(deployment *appsv1.Deployment) {
	key, err := cache.MetaNamespaceKeyFunc(deployment)
	if err != nil {
		log.Printf("❌ Error creating key for deployment: %v", err)
		return
	}

	// Update local cache
	e.deploymentCache[key] = deployment.DeepCopy()
	e.workqueue.Add(fmt.Sprintf("add:%s", key))

	replicas := int32(0)
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	}
	log.Printf("🆕 New deployment detected: %s/%s (replicas: %d)",
		deployment.Namespace, deployment.Name, replicas)
}

func (e *EventProcessor) handleUpdateEvent(oldDeployment, newDeployment *appsv1.Deployment) {
	if !e.config.CustomLogic.EnableUpdateHandling {
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(newDeployment)
	if err != nil {
		log.Printf("❌ Error creating key for deployment: %v", err)
		return
	}

	// Update local cache
	e.deploymentCache[key] = newDeployment.DeepCopy()

	if e.hasSignificantChanges(oldDeployment, newDeployment) {
		e.workqueue.Add(fmt.Sprintf("update:%s", key))
		e.processDeploymentUpdate(oldDeployment, newDeployment)
	}
}

func (e *EventProcessor) handleDeleteEvent(deployment *appsv1.Deployment) {
	if !e.config.CustomLogic.EnableDeleteHandling {
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(deployment)
	if err != nil {
		log.Printf("❌ Error creating key for deployment: %v", err)
		return
	}

	// Remove from local cache
	delete(e.deploymentCache, key)
	e.workqueue.Add(fmt.Sprintf("delete:%s", key))
	e.processDeploymentDeletion(deployment)
}

// Step 7+: Logic to detect significant changes
func (e *EventProcessor) hasSignificantChanges(old, new *appsv1.Deployment) bool {
	// Check replica count changes
	if old.Spec.Replicas != nil && new.Spec.Replicas != nil {
		if *old.Spec.Replicas != *new.Spec.Replicas {
			return true
		}
	}

	// Check image changes
	if len(old.Spec.Template.Spec.Containers) > 0 && len(new.Spec.Template.Spec.Containers) > 0 {
		if old.Spec.Template.Spec.Containers[0].Image != new.Spec.Template.Spec.Containers[0].Image {
			return true
		}
	}

	// Check generation changes (spec updates)
	if old.Generation != new.Generation {
		return true
	}

	return false
}

func (e *EventProcessor) processDeploymentUpdate(old, new *appsv1.Deployment) {
	log.Printf("🔧 Processing update for deployment %s/%s", new.Namespace, new.Name)

	// Check replica scaling
	if old.Spec.Replicas != nil && new.Spec.Replicas != nil {
		oldReplicas := *old.Spec.Replicas
		newReplicas := *new.Spec.Replicas

		if oldReplicas != newReplicas {
			if newReplicas > oldReplicas {
				log.Printf("📈 SCALE UP: %s/%s scaled from %d to %d replicas",
					new.Namespace, new.Name, oldReplicas, newReplicas)
			} else {
				log.Printf("📉 SCALE DOWN: %s/%s scaled from %d to %d replicas",
					new.Namespace, new.Name, oldReplicas, newReplicas)
			}
		}
	}

	// Check image updates
	if len(old.Spec.Template.Spec.Containers) > 0 && len(new.Spec.Template.Spec.Containers) > 0 {
		oldImage := old.Spec.Template.Spec.Containers[0].Image
		newImage := new.Spec.Template.Spec.Containers[0].Image

		if oldImage != newImage {
			log.Printf("🔄 IMAGE UPDATE: %s/%s image changed from %s to %s",
				new.Namespace, new.Name, oldImage, newImage)
		}
	}

	// Check deployment status
	if new.Status.ReadyReplicas != new.Status.Replicas {
		log.Printf("⚠️ UNHEALTHY: %s/%s has %d/%d replicas ready",
			new.Namespace, new.Name, new.Status.ReadyReplicas, new.Status.Replicas)
	}
}

func (e *EventProcessor) processDeploymentDeletion(deployment *appsv1.Deployment) {
	log.Printf("🗑️ Processing deletion for deployment %s/%s", deployment.Namespace, deployment.Name)

	if deployment.Spec.Replicas != nil {
		log.Printf("📊 Deleted deployment had %d replicas", *deployment.Spec.Replicas)
	}

	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		log.Printf("🐳 Deleted deployment was running image: %s",
			deployment.Spec.Template.Spec.Containers[0].Image)
	}
}

func (e *EventProcessor) runWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			obj, shutdown := e.workqueue.Get()
			if shutdown {
				return
			}

			// Process the work item
			if objStr, ok := obj.(string); ok {
				log.Printf("🔄 Processing work item: %s", objStr)
			}

			e.workqueue.Done(obj)
		}
	}
}

// Step 7++: Configuration loading for informers
func loadInformerConfig() (*InformerConfig, error) {
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

	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err == nil {
			log.Printf("📄 Using config file: %s", viper.ConfigFileUsed())
			if err := viper.Unmarshal(config); err != nil {
				return nil, fmt.Errorf("error unmarshaling config: %v", err)
			}
		} else {
			log.Printf("⚠️ Could not read config file %s: %v", configFile, err)
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

	return config, nil
}

// Step 7: Watch command with informers
var watchInformerCmd = &cobra.Command{
	Use:   "watch-informer",
	Short: "Watch deployment events using k8s.io/client-go informers (Step 7)",
	Long: `Start watching Kubernetes deployment events using k8s.io/client-go informers with custom event handling.

Step 7 Features:
• Uses k8s.io/client-go SharedInformerFactory for list/watch operations
• Supports both kubeconfig and in-cluster authentication  
• Reports all deployment events (ADD/UPDATE/DELETE) in logs
• Custom logic for processing significant deployment changes
• Configurable resync period and worker count
• Cache storage for deployment resources

Authentication:
• Default: kubeconfig from ~/.kube/config
• In-cluster: use --in-cluster flag when running in pod`,
	Run: func(cmd *cobra.Command, args []string) {
		runWatchInformer()
	},
}

func runWatchInformer() {
	log.Println("🎯 Starting k8s-cli deployment watcher with k8s.io/client-go informers...")

	config, err := loadInformerConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	log.Printf("⚙️ Configuration loaded - ResyncPeriod: %v, Workers: %d, LogEvents: %v",
		config.ResyncPeriod, config.Workers, config.LogEvents)

	// Step 7: Get Kubernetes client with kubeconfig or in-cluster auth
	clientset, err := GetKubernetesClient()
	if err != nil {
		log.Fatalf("❌ Failed to create Kubernetes client: %v", err)
	}

	// Test connection to cluster
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		log.Fatalf("❌ Failed to connect to Kubernetes cluster: %v", err)
	}
	log.Printf("✅ Successfully connected to Kubernetes cluster (version: %s)", serverVersion.String())

	processor := NewEventProcessor(clientset, config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := processor.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start event processor: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("🎉 k8s-cli is now watching deployment events using informers. Press Ctrl+C to stop.")
	log.Println("📋 Step 7 Features Active:")
	log.Println("   ✅ k8s.io/client-go SharedInformerFactory")
	log.Println("   ✅ list/watch informer for deployment resources")
	log.Println("   ✅ kubeconfig/in-cluster authentication")
	log.Println("   ✅ Event logging (ADD/UPDATE/DELETE)")
	log.Println("   ✅ Custom logic for deployment changes")
	log.Println("   ✅ Cache storage for informer data")

	<-signalChan
	log.Println("\n🛑 Shutdown signal received, stopping...")

	processor.Stop()
	cancel()

	log.Println("👋 k8s-cli stopped gracefully")
}

func init() {
	// Add flags for Step 7
	watchInformerCmd.Flags().DurationVar(&informerResyncPeriod, "resync-period", 0, "Informer resync period")
	watchInformerCmd.Flags().IntVar(&informerWorkers, "workers", 0, "Number of worker goroutines")
	watchInformerCmd.Flags().BoolVar(&enableEventLogging, "log-events", true, "Enable event logging")
	watchInformerCmd.Flags().StringVar(&configFile, "config", "", "Path to configuration file")

	// Register command
	RootCmd.AddCommand(watchInformerCmd)
}
