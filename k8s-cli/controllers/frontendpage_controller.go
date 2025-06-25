package controllers

import (
	"context"
	"fmt"
	"log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	k8scliv1 "github.com/dereban25/go-kubernetes-controllers/k8s-cli/api/v1"
)

// FrontendPageReconciler reconciles a FrontendPage object
type FrontendPageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=k8scli.dev,resources=frontendpages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8scli.dev,resources=frontendpages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8scli.dev,resources=frontendpages/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *FrontendPageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Printf("🔄 Step 11: Reconciling FrontendPage %s/%s", req.Namespace, req.Name)

	// Fetch the FrontendPage instance
	var frontendPage k8scliv1.FrontendPage
	if err := r.Get(ctx, req.NamespacedName, &frontendPage); err != nil {
		if errors.IsNotFound(err) {
			log.Printf("🗑️ Step 11: FrontendPage %s/%s not found, probably deleted", req.Namespace, req.Name)
			return ctrl.Result{}, nil
		}
		log.Printf("❌ Error fetching FrontendPage: %v", err)
		return ctrl.Result{}, err
	}

	log.Printf("📊 Step 11: FrontendPage Details:")
	log.Printf("   Title: %s", frontendPage.Spec.Title)
	log.Printf("   Description: %s", frontendPage.Spec.Description)
	log.Printf("   Path: %s", frontendPage.Spec.Path)
	log.Printf("   Template: %s", frontendPage.Spec.Template)
	log.Printf("   Replicas: %d", frontendPage.Spec.Replicas)
	log.Printf("   Image: %s", frontendPage.Spec.Image)

	// Update status phase
	if frontendPage.Status.Phase == "" {
		frontendPage.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, &frontendPage); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Create or update deployment
	deployment, err := r.createOrUpdateDeployment(ctx, &frontendPage)
	if err != nil {
		log.Printf("❌ Step 11: Failed to create/update deployment: %v", err)
		r.updateStatus(ctx, &frontendPage, "Failed", false, err.Error())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Create or update service
	service, err := r.createOrUpdateService(ctx, &frontendPage)
	if err != nil {
		log.Printf("❌ Step 11: Failed to create/update service: %v", err)
		r.updateStatus(ctx, &frontendPage, "Failed", false, err.Error())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Check deployment readiness
	ready := deployment.Status.ReadyReplicas == deployment.Status.Replicas && deployment.Status.Replicas > 0

	// Update status
	phase := "Running"
	if !ready {
		phase = "Pending"
	}

	url := fmt.Sprintf("http://%s.%s.svc.cluster.local%s", service.Name, service.Namespace, frontendPage.Spec.Path)

	frontendPage.Status.Phase = phase
	frontendPage.Status.Ready = ready
	frontendPage.Status.URL = url
	frontendPage.Status.DeploymentName = deployment.Name
	frontendPage.Status.ServiceName = service.Name
	frontendPage.Status.LastUpdated = time.Now().Format(time.RFC3339)
	frontendPage.Status.ObservedGeneration = frontendPage.Generation

	if ready {
		frontendPage.Status.Message = fmt.Sprintf("Deployment %s is ready", deployment.Name)
	} else {
		frontendPage.Status.Message = fmt.Sprintf("Deployment %s is not ready yet", deployment.Name)
	}

	if err := r.Status().Update(ctx, &frontendPage); err != nil {
		return ctrl.Result{}, err
	}

	if ready {
		log.Printf("✅ Step 11: FrontendPage %s/%s is ready at %s", req.Namespace, req.Name, url)
	} else {
		log.Printf("⏳ Step 11: FrontendPage %s/%s is not ready yet, requeuing...", req.Namespace, req.Name)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	log.Printf("🎯 Step 11: Reconciliation completed for FrontendPage %s/%s", req.Namespace, req.Name)
	return ctrl.Result{}, nil
}

func (r *FrontendPageReconciler) createOrUpdateDeployment(ctx context.Context, frontendPage *k8scliv1.FrontendPage) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      frontendPage.Name + "-deployment",
			Namespace: frontendPage.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		// Set owner reference
		if err := controllerutil.SetControllerReference(frontendPage, deployment, r.Scheme); err != nil {
			return err
		}

		// Configure deployment spec
		replicas := frontendPage.Spec.Replicas
		if replicas == 0 {
			replicas = 1
		}

		image := frontendPage.Spec.Image
		if image == "" {
			image = "nginx:1.20"
		}

		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":          frontendPage.Name,
					"frontendpage": frontendPage.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":          frontendPage.Name,
						"frontendpage": frontendPage.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "frontend",
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Name:          "http",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "FRONTEND_TITLE",
									Value: frontendPage.Spec.Title,
								},
								{
									Name:  "FRONTEND_DESCRIPTION",
									Value: frontendPage.Spec.Description,
								},
								{
									Name:  "FRONTEND_PATH",
									Value: frontendPage.Spec.Path,
								},
							},
						},
					},
				},
			},
		}

		// Add config as environment variables
		for key, value := range frontendPage.Spec.Config {
			deployment.Spec.Template.Spec.Containers[0].Env = append(
				deployment.Spec.Template.Spec.Containers[0].Env,
				corev1.EnvVar{Name: key, Value: value},
			)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Printf("🔨 Step 11: Deployment %s %s", deployment.Name, op)
	return deployment, nil
}

func (r *FrontendPageReconciler) createOrUpdateService(ctx context.Context, frontendPage *k8scliv1.FrontendPage) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      frontendPage.Name + "-service",
			Namespace: frontendPage.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		// Set owner reference
		if err := controllerutil.SetControllerReference(frontendPage, service, r.Scheme); err != nil {
			return err
		}

		service.Spec = corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app":          frontendPage.Name,
				"frontendpage": frontendPage.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Printf("🔨 Step 11: Service %s %s", service.Name, op)
	return service, nil
}

func (r *FrontendPageReconciler) updateStatus(ctx context.Context, frontendPage *k8scliv1.FrontendPage, phase string, ready bool, message string) {
	frontendPage.Status.Phase = phase
	frontendPage.Status.Ready = ready
	frontendPage.Status.LastUpdated = time.Now().Format(time.RFC3339)
	frontendPage.Status.Message = message
	frontendPage.Status.ObservedGeneration = frontendPage.Generation
	r.Status().Update(ctx, frontendPage)
}

// SetupWithManager sets up the controller with the Manager.
func (r *FrontendPageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8scliv1.FrontendPage{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
