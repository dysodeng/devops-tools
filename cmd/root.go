package cmd

import (
	"fmt"
	"github.com/dysodeng/devops-tools/internal/module"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:     "devops",
	Short:   "运维工具箱",
	Long:    "运维工具箱",
	Version: fmt.Sprintf("%s\n", module.Version()),
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(module.VersionCmd)
	rootCmd.AddCommand(module.SystemCmd)
	rootCmd.AddCommand(module.ContainerCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}
}
