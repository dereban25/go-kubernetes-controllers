package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PrintPods выводит список подов в указанном формате
func PrintPods(pods []corev1.Pod, format string) error {
	switch format {
	case "json":
		printPodsJSON(pods)
	case "yaml":
		printPodsYAML(pods)
	default:
		printPodsTable(pods)
	}
	return nil
}

// PrintDeployments выводит список деплойментов в указанном формате
func PrintDeployments(deployments []appsv1.Deployment, format string) error {
	switch format {
	case "json":
		printDeploymentsJSON(deployments)
	case "yaml":
		printDeploymentsYAML(deployments)
	default:
		printDeploymentsTable(deployments)
	}
	return nil
}

// PrintServices выводит список сервисов в указанном формате
func PrintServices(services []corev1.Service, format string) error {
	switch format {
	case "json":
		printServicesJSON(services)
	case "yaml":
		printServicesYAML(services)
	default:
		printServicesTable(services)
	}
	return nil
}

func printPodsTable(pods []corev1.Pod) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "NAMESPACE", "STATUS", "READY", "RESTARTS", "AGE"})

	for _, pod := range pods {
		ready := fmt.Sprintf("%d/%d", countReadyContainers(pod), len(pod.Spec.Containers))
		restarts := fmt.Sprintf("%d", countRestarts(pod))
		age := formatAge(pod.CreationTimestamp)

		table.Append([]string{
			pod.Name,
			pod.Namespace,
			string(pod.Status.Phase),
			ready,
			restarts,
			age,
		})
	}

	table.Render()
}

func printDeploymentsTable(deployments []appsv1.Deployment) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "NAMESPACE", "READY", "UP-TO-DATE", "AVAILABLE", "AGE"})

	for _, deployment := range deployments {
		replicas := int32(0)
		if deployment.Spec.Replicas != nil {
			replicas = *deployment.Spec.Replicas
		}
		ready := fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, replicas)
		upToDate := fmt.Sprintf("%d", deployment.Status.UpdatedReplicas)
		available := fmt.Sprintf("%d", deployment.Status.AvailableReplicas)
		age := formatAge(deployment.CreationTimestamp)

		table.Append([]string{
			deployment.Name,
			deployment.Namespace,
			ready,
			upToDate,
			available,
			age,
		})
	}

	table.Render()
}

func printServicesTable(services []corev1.Service) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "NAMESPACE", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORT(S)", "AGE"})

	for _, service := range services {
		serviceType := string(service.Spec.Type)
		clusterIP := service.Spec.ClusterIP
		externalIP := getExternalIP(service)
		ports := getPorts(service)
		age := formatAge(service.CreationTimestamp)

		table.Append([]string{
			service.Name,
			service.Namespace,
			serviceType,
			clusterIP,
			externalIP,
			ports,
			age,
		})
	}

	table.Render()
}

func printPodsJSON(pods []corev1.Pod) {
	data, err := json.MarshalIndent(pods, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling pods to JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func printDeploymentsJSON(deployments []appsv1.Deployment) {
	data, err := json.MarshalIndent(deployments, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling deployments to JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func printServicesJSON(services []corev1.Service) {
	data, err := json.MarshalIndent(services, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling services to JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func printPodsYAML(pods []corev1.Pod) {
	fmt.Println("# Pods YAML output")
	printPodsJSON(pods)
}

func printDeploymentsYAML(deployments []appsv1.Deployment) {
	fmt.Println("# Deployments YAML output")
	printDeploymentsJSON(deployments)
}

func printServicesYAML(services []corev1.Service) {
	fmt.Println("# Services YAML output")
	printServicesJSON(services)
}

// Вспомогательные функции
func countReadyContainers(pod corev1.Pod) int32 {
	var ready int32
	for _, status := range pod.Status.ContainerStatuses {
		if status.Ready {
			ready++
		}
	}
	return ready
}

func countRestarts(pod corev1.Pod) int32 {
	var restarts int32
	for _, status := range pod.Status.ContainerStatuses {
		restarts += status.RestartCount
	}
	return restarts
}

func formatAge(timestamp metav1.Time) string {
	duration := time.Since(timestamp.Time)

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd", days)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

func getExternalIP(service corev1.Service) string {
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		if service.Status.LoadBalancer.Ingress[0].IP != "" {
			return service.Status.LoadBalancer.Ingress[0].IP
		}
		if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
			return service.Status.LoadBalancer.Ingress[0].Hostname
		}
	}

	if len(service.Spec.ExternalIPs) > 0 {
		return strings.Join(service.Spec.ExternalIPs, ",")
	}

	return "<none>"
}

func getPorts(service corev1.Service) string {
	var ports []string
	for _, port := range service.Spec.Ports {
		if port.NodePort != 0 {
			ports = append(ports, fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol))
		} else {
			ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
		}
	}
	return strings.Join(ports, ",")
}
