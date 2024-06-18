package cmd

import (
	"fmt"
	"log"
	"os"

	_ "github.com/dysodeng/devops-tools/internal/module"
	"github.com/dysodeng/devops-tools/internal/module/container"
	"github.com/dysodeng/devops-tools/internal/module/kubernetes"
	"github.com/dysodeng/devops-tools/internal/module/system"
	"github.com/dysodeng/devops-tools/internal/module/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "devops",
	Short:   "运维工具箱",
	Long:    "运维工具箱",
	Version: fmt.Sprintf("%s\n", version.Version()),
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(version.Cmd)
	rootCmd.AddCommand(system.Cmd)
	rootCmd.AddCommand(container.Cmd)
	rootCmd.AddCommand(kubernetes.Cmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}
}
