package module

import (
	"fmt"
	"github.com/spf13/cobra"
)

const version = "v0.0.1"

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "version for devops",
	Long:  "version for devops",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("devops version %s\n", version)
	},
}

func Version() string {
	return version
}
