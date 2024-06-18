package kubernetes

import "github.com/spf13/cobra"

// Cmd k8s配置命令
var Cmd = &cobra.Command{
	Use:   "k8s",
	Short: "Kubernetes配置",
	Long:  "Kubernetes配置",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func InitKubernetesCmd() {
	loadImageCmd.Flags().BoolVarP(&containerWithDocker, "with-docker", "", false, "使用Docker，默认为containerd")
	installKubernetesCmd.Flags().BoolVarP(&containerWithDocker, "with-docker", "", false, "使用Docker，默认为containerd")
	installKubernetesCmd.Flags().StringVarP(&withKubernetesVersion, "with-version", "", "v1.27.6", "指定Kubernetes版本")
	initKubernetesClusterCmd.Flags().StringVarP(&withKubernetesVersion, "with-version", "", "v1.27.6", "指定Kubernetes版本")
	joinKubernetesNodeCmd.Flags().BoolVarP(&joinMasterNode, "with-master", "", false, "加入master节点")
	Cmd.AddCommand(loadImageCmd, installKubernetesCmd, initKubernetesClusterCmd, joinKubernetesNodeCmd)
}
