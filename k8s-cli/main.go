package main

import (
	"github.com/dereban25/go-kubernetes-controllers/k8s-cli/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
