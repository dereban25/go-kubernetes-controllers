package cmd

import (
	"fmt"
	"github.com/dereban25/go-kubernetes-controllers/k8s-cli/internal/k8s"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// contextCmd represents the context command
var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage Kubernetes contexts",
	Long:  "Commands for working with Kubernetes contexts - viewing, switching",
}

// contextListCmd lists contexts
var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all contexts",
	Long:  "Show all available Kubernetes contexts",
	Example: `  # Show all contexts
  k8s-cli context list`,
	RunE: runContextList,
}

// contextCurrentCmd shows current context
var contextCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current context",
	Long:  "Show current active Kubernetes context",
	Example: `  # Show current context
  k8s-cli context current`,
	RunE: runContextCurrent,
}

// contextSetCmd switches context
var contextSetCmd = &cobra.Command{
	Use:   "set <context-name>",
	Short: "Switch context",
	Long:  "Switch to specified Kubernetes context",
	Args:  cobra.ExactArgs(1),
	Example: `  # Switch to context
  k8s-cli context set my-cluster`,
	RunE: runContextSet,
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextCurrentCmd)
	contextCmd.AddCommand(contextSetCmd)
}

func runContextList(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	contexts, err := client.GetContexts()
	if err != nil {
		return fmt.Errorf("error getting contexts: %w", err)
	}

	currentContext, err := client.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("error getting current context: %w", err)
	}

	fmt.Println("Available contexts:")
	for _, context := range contexts {
		if context == currentContext {
			fmt.Printf("* %s (current)\n", context)
		} else {
			fmt.Printf("  %s\n", context)
		}
	}

	return nil
}

func runContextCurrent(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	currentContext, err := client.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("error getting current context: %w", err)
	}

	fmt.Printf("Current context: %s\n", currentContext)
	return nil
}

func runContextSet(cmd *cobra.Command, args []string) error {
	contextName := args[0]
	kubeconfigPath := viper.GetString("kubeconfig")

	client, err := k8s.NewClient(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	err = client.SetContext(contextName, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("error switching context: %w", err)
	}

	fmt.Printf("Context switched to: %s\n", contextName)
	return nil
}
