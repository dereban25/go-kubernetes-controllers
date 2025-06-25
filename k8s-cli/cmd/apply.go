package cmd

import (
	"fmt"
	"github.com/dereban25/go-kubernetes-controllers/k8s-cli/internal/k8s"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Create resources from YAML file",
	Long:  "Create or update Kubernetes resources from YAML file",
}

// applyFileCmd creates resources from file
var applyFileCmd = &cobra.Command{
	Use:   "file <filename>",
	Short: "Apply YAML file",
	Long:  "Create or update Kubernetes resources from specified YAML file",
	Args:  cobra.ExactArgs(1),
	Example: `  # Apply YAML file
  k8s-cli apply file pod.yaml

  # Apply file in specific namespace
  k8s-cli apply file deployment.yaml -n my-app`,
	RunE: runApplyFile,
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.AddCommand(applyFileCmd)
}

func runApplyFile(cmd *cobra.Command, args []string) error {
	filename := args[0]

	// Read YAML file
	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	// Create Kubernetes client
	client, err := k8s.NewClient(viper.GetString("kubeconfig"))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	namespace := viper.GetString("namespace")

	// Apply YAML
	err = client.CreateFromYAML(yamlData, namespace)
	if err != nil {
		return fmt.Errorf("error applying YAML: %w", err)
	}

	fmt.Printf("✅ Resources successfully created from file: %s\n", filename)
	return nil
}
