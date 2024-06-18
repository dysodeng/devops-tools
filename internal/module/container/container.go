package container

import (
	"fmt"
	"os"

	"github.com/dysodeng/devops-tools/internal/module/system"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "container",
	Short: "容器运行时配置",
	Long:  "容器运行时配置",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// containerWithDocker 使用Docker，否则使用containerd
var containerWithDocker bool

// containerWithDataDirectory 指定容器运行时数据存储目录
var containerWithDataDirectory string

// installContainerCmd 安装容器运行时
var installContainerCmd = &cobra.Command{
	Use:   "install",
	Short: "安装容器运行时，默认安装containerd",
	Long:  "安装容器运行时，默认安装containerd",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if containerWithDocker {
			err = installDocker(system.System.LinuxDistro, system.System.Arch)
		} else {
			err = installContainerd(system.System.LinuxDistro, system.System.Arch)
		}
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

func InitContainerCmd() {
	installContainerCmd.Flags().BoolVarP(&containerWithDocker, "with-docker", "", false, "安装Docker")
	installContainerCmd.Flags().StringVarP(&containerWithDataDirectory, "with-data", "", "", "指定容器运行时数据存储目录")
	Cmd.AddCommand(installContainerCmd)
}
