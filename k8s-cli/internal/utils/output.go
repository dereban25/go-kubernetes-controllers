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
	"sigs.k8s.io/yaml"
)

// PrintPods выводит список подов
func PrintPods(pods []corev1.Pod, format string) error {
	switch format {
	case "json":
		return printJSON(pods)
	case "yaml":
		return printYAML(pods)
	default:
		return printPodsTable(pods)
	}
}

// PrintDeployments выводит список деплойментов
func PrintDeployments(deployments []appsv1.Deployment, format string) error {
	switch format {
	case "json":
		return printJSON(deployments)
	case "yaml":
		return printYAML(deployments)
	default:
		return printDeploymentsTable(deployments)
	}
}

// PrintServices выводит список сервисов
func PrintServices(services []corev1.Service, format string) error {
	switch format {
	case "json":
		return printJSON(services)
	case "yaml":
		return printYAML(services)
	default:
		return printServicesTable(services)
	}
}

func printPodsTable(pods []corev1.Pod) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	for _, pod := range pods {
		ready := fmt.Sprintf("%d/%d", getReadyContainers(pod), len(pod.Spec.Containers))
		status := string(pod.Status.Phase)
		restarts := getRestartCount(pod)
		age := getAge(pod.CreationTimestamp)

		table.Append([]string{
			pod.Name,
			ready,
			status,
			fmt.Sprintf("%d", restarts),
			age,
		})
	}

	table.Render()
	return nil
}

func printDeploymentsTable(deployments []appsv1.Deployment) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "READY", "UP-TO-DATE", "AVAILABLE", "AGE"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	for _, deployment := range deployments {
		replicas := int32(0)
		if deployment.Spec.Replicas != nil {
			replicas = *deployment.Spec.Replicas
		}
		ready := fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, replicas)
		upToDate := fmt.Sprintf("%d", deployment.Status.UpdatedReplicas)
		available := fmt.Sprintf("%d", deployment.Status.AvailableReplicas)
		age := getAge(deployment.CreationTimestamp)

		table.Append([]string{
			deployment.Name,
			ready,
			upToDate,
			available,
			age,
		})
	}

	table.Render()
	return nil
}

func printServicesTable(services []corev1.Service) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORT(S)", "AGE"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	for _, service := range services {
		svcType := string(service.Spec.Type)
		clusterIP := service.Spec.ClusterIP
		externalIP := getExternalIP(service)
		ports := getPorts(service)
		age := getAge(service.CreationTimestamp)

		table.Append([]string{
			service.Name,
			svcType,
			clusterIP,
			externalIP,
			ports,
			age,
		})
	}

	table.Render()
	return nil
}

func printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func printYAML(data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Print(string(yamlData))
	return nil
}

// Вспомогательные функции
func getReadyContainers(pod corev1.Pod) int {
	ready := 0
	for _, status := range pod.Status.ContainerStatuses {
		if status.Ready {
			ready++
		}
	}
	return ready
}

func getRestartCount(pod corev1.Pod) int32 {
	var restarts int32
	for _, status := range pod.Status.ContainerStatuses {
		restarts += status.RestartCount
	}
	return restarts
}

func getAge(timestamp metav1.Time) string {
	age := time.Since(timestamp.Time)
	if age.Hours() >= 24 {
		return fmt.Sprintf("%dd", int(age.Hours()/24))
	} else if age.Hours() >= 1 {
		return fmt.Sprintf("%dh", int(age.Hours()))
	} else if age.Minutes() >= 1 {
		return fmt.Sprintf("%dm", int(age.Minutes()))
	} else {
		return fmt.Sprintf("%ds", int(age.Seconds()))
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
		return service.Spec.ExternalIPs[0]
	}

	return "<none>"
}

func getPorts(service corev1.Service) string {
	if len(service.Spec.Ports) == 0 {
		return "<none>"
	}

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
