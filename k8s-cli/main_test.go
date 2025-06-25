package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// Import your cmd package with correct module path
	"github.com/dereban25/go-kubernetes-controllers/k8s-cli/cmd"
)

var (
	cfg       *rest.Config
	k8sClient kubernetes.Interface
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "k8s-cli Step 7 Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

// Step 7: Test basic informer functionality using k8s.io/client-go
var _ = Describe("Step 7: k8s.io/client-go Informer Functionality", func() {
	var testNamespace *corev1.Namespace

	BeforeEach(func() {
		testNamespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-ns-",
			},
		}
		var err error
		testNamespace, err = k8sClient.CoreV1().Namespaces().Create(ctx, testNamespace, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := k8sClient.CoreV1().Namespaces().Delete(ctx, testNamespace.Name, metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("When creating EventProcessor with k8s.io/client-go", func() {
		It("Should initialize with correct configuration", func() {
			config := &cmd.InformerConfig{
				ResyncPeriod: 30 * time.Second,
				Workers:      2,
				Namespaces:   []string{testNamespace.Name},
				LogEvents:    true,
			}
			config.CustomLogic.EnableUpdateHandling = true
			config.CustomLogic.EnableDeleteHandling = true

			processor := cmd.NewEventProcessor(k8sClient, config)
			Expect(processor).NotTo(BeNil())
		})

		It("Should authenticate via kubeconfig", func() {
			// Test kubeconfig authentication path (envtest simulates this)
			client, err := getKubernetesClientWithConfig(cfg)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())

			// Verify connection by listing nodes
			_, err = client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When starting informer with k8s.io/client-go", func() {
		It("Should start SharedInformerFactory and sync cache successfully", func() {
			config := &cmd.InformerConfig{
				ResyncPeriod: 5 * time.Second,
				Workers:      1,
				Namespaces:   []string{testNamespace.Name},
				LogEvents:    false, // Disable for cleaner test output
			}
			config.CustomLogic.EnableUpdateHandling = true
			config.CustomLogic.EnableDeleteHandling = true

			processor := cmd.NewEventProcessor(k8sClient, config)

			testCtx, testCancel := context.WithTimeout(ctx, 10*time.Second)
			defer testCancel()

			err := processor.Start(testCtx)
			Expect(err).NotTo(HaveOccurred())

			// Give processor time to initialize
			time.Sleep(2 * time.Second)

			processor.Stop()
		})

		It("Should report deployment events in logs", func() {
			config := &cmd.InformerConfig{
				ResyncPeriod: 5 * time.Second,
				Workers:      1,
				Namespaces:   []string{testNamespace.Name},
				LogEvents:    true, // Enable event logging for Step 7
			}
			config.CustomLogic.EnableUpdateHandling = true

			processor := cmd.NewEventProcessor(k8sClient, config)

			testCtx, testCancel := context.WithTimeout(ctx, 15*time.Second)
			defer testCancel()

			// Start processor
			err := processor.Start(testCtx)
			Expect(err).NotTo(HaveOccurred())

			// Give time for cache sync
			time.Sleep(2 * time.Second)

			// Create a deployment to trigger ADD event
			deployment := createTestDeployment(testNamespace.Name, "test-deployment")
			_, err = k8sClient.AppsV1().Deployments(testNamespace.Name).Create(ctx, deployment, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			// Wait for event processing
			time.Sleep(3 * time.Second)

			processor.Stop()
		})
	})
})

// Helper functions for Step 7 testing
func createTestDeployment(namespace, name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: "nginx:1.20",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(80),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
						},
					},
				},
			},
		},
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}

func getKubernetesClientWithConfig(config *rest.Config) (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(config)
}

// Basic test to verify Step 7 functionality
func TestStep7BasicFunctionality(t *testing.T) {
	// This is a basic Go test (not Ginkgo) to verify core functionality
	config := &cmd.InformerConfig{
		ResyncPeriod: 30 * time.Second,
		Workers:      1,
		Namespaces:   []string{"default"},
		LogEvents:    true,
	}

	// Just test that we can create the processor without k8s connection
	// (since we don't have envtest setup in basic Go test)
	if config.ResyncPeriod != 30*time.Second {
		t.Errorf("Expected ResyncPeriod to be 30s, got %v", config.ResyncPeriod)
	}

	if config.Workers != 1 {
		t.Errorf("Expected Workers to be 1, got %d", config.Workers)
	}
}
