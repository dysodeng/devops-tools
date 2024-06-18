package kubernetes

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dysodeng/devops-tools/internal/module/system"
	"github.com/dysodeng/devops-tools/internal/pkg"
	"github.com/spf13/cobra"
)

// withKubernetesVersion k8s版本
var withKubernetesVersion string

// installKubernetesCmd 安装k8s组件命令
var installKubernetesCmd = &cobra.Command{
	Use:   "install",
	Short: "安装Kubernetes组件",
	Long:  "安装Kubernetes组件",
	Run: func(cmd *cobra.Command, args []string) {
		if err := installKubernetes(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

// installKubernetes 安装k8s组件
func installKubernetes() error {
	switch system.System.LinuxDistro {
	case "CentOS":
		// 禁用swap分区
		_ = pkg.ExecCmd(exec.Command("swapoff", "-a"))
		_ = pkg.ExecCmd(exec.Command("sed", "-i", "s/.*swap.*/#&/", "/etc/fstab"))

		// k8s sysctl 配置
		if err := k8sSysctlConfig(); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("sysctl", "--system")); err != nil {
			return err
		}

		// 配置ipvsadm
		if err := pkg.ExecCmd(exec.Command("yum", "install", "-y", "ipset", "ipvsadm")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("modprobe", "overlay")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("modprobe", "br_netfilter")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("modprobe", "ip_tables")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("modprobe", "iptable_filter")); err != nil {
			return err
		}
		if err := k8sModuleLoadConfig(); err != nil {
			return err
		}

		// 安装k8s组件
		if err := k8sRepoCentosConfig(); err != nil {
			return err
		}

		k8sVersion := strings.Replace(withKubernetesVersion, "v", "", -1)
		if err := pkg.ExecCmd(
			exec.Command(
				"yum",
				"install",
				"-y",
				"kubelet-"+k8sVersion+"-0",
				"kubeadm-"+k8sVersion+"-0",
				"kubectl-"+k8sVersion+"-0",
				"kubernetes-cni",
				"--disableexcludes=kubernetes",
				"--nogpgcheck",
			),
		); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("systemctl", "daemon-reload")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("systemctl", "enable", "kubelet", "--now")); err != nil {
			return err
		}
		_ = pkg.ExecCmd(exec.Command("systemctl", "status", "kubelet"))
		break

	case "Ubuntu":
		// 禁用swap分区
		_ = pkg.ExecCmd(exec.Command("swapoff", "-a"))
		_ = pkg.ExecCmd(exec.Command("sed", "-i", "s/.*swap.*/#&/", "/etc/fstab"))

		if err := pkg.ExecCmd(exec.Command("apt", "update")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("apt", "install", "-y", "apt-transport-https")); err != nil {
			return err
		}

		// k8s sysctl 配置
		if err := k8sSysctlConfig(); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("sysctl", "--system")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("apt", "install", "-y", "ipset", "ipvsadm")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("modprobe", "overlay")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("modprobe", "br_netfilter")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("modprobe", "ip_tables")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("modprobe", "iptable_filter")); err != nil {
			return err
		}

		// 安装k8s组件
		if err := pkg.ExecCmd(
			exec.Command(
				"/bin/bash",
				"-c",
				`curl -fsSL https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add -`,
			),
		); err != nil {
			return err
		}
		if pkg.CheckNetworkFileExists(
			"https://mirrors.aliyun.com/kubernetes/apt/kubernetes-" + system.System.CodeName + "/Release",
		) {
			if err := pkg.ExecCmd(
				exec.Command(
					"/bin/bash",
					"-c",
					fmt.Sprintf(
						`add-apt-repository "deb [arch=%s] https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-%s main")`,
						system.System.Arch,
						system.System.CodeName,
					),
				),
			); err != nil {
				return err
			}
		} else {
			if err := pkg.ExecCmd(
				exec.Command(
					"/bin/bash",
					"-c",
					fmt.Sprintf(
						`add-apt-repository "deb [arch=%s] https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main"`,
						system.System.Arch,
					),
				),
			); err != nil {
				return err
			}
		}
		if err := pkg.ExecCmd(exec.Command("apt", "update")); err != nil {
			return err
		}
		k8sVersion := strings.Replace(withKubernetesVersion, "v", "", -1)
		if err := pkg.ExecCmd(
			exec.Command(
				"apt",
				"install",
				"-y",
				"kubelet="+k8sVersion+"-00",
				"kubeadm="+k8sVersion+"-00",
				"kubectl="+k8sVersion+"-00",
				"kubernetes-cni",
			),
		); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("systemctl", "daemon-reload")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("systemctl", "enable", "kubelet", "--now")); err != nil {
			return err
		}
		_ = pkg.ExecCmd(exec.Command("systemctl", "status", "kubelet"))
		break

	case "Debian":
		// 禁用swap分区
		_ = pkg.ExecCmd(exec.Command("swapoff", "-a"))
		_ = pkg.ExecCmd(exec.Command("sed", "-i", "s/.*swap.*/#&/", "/etc/fstab"))

		if err := pkg.ExecCmd(exec.Command("apt", "update")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("apt", "install", "-y", "apt-transport-https", "gnupg2", "gnupg1", "gnupg")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(
			exec.Command(
				"apt",
				"install",
				"-y",
				"software-properties-common",
				"dirmngr",
				"ca-certificates",
			),
		); err != nil {
			return err
		}

		// k8s sysctl 配置
		if err := k8sSysctlConfig(); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("sysctl", "--system")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("apt", "install", "-y", "ipset", "ipvsadm")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("modprobe", "overlay")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("modprobe", "br_netfilter")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("modprobe", "ip_tables")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("modprobe", "iptable_filter")); err != nil {
			return err
		}

		// 安装k8s组件
		if err := pkg.ExecCmd(
			exec.Command(
				"/bin/bash",
				"-c",
				`curl -fsSL https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add -`,
			),
		); err != nil {
			return err
		}
		if pkg.CheckNetworkFileExists(
			"https://mirrors.aliyun.com/kubernetes/apt/kubernetes-" + system.System.CodeName + "/Release",
		) {
			if err := pkg.ExecCmd(
				exec.Command(
					"/bin/bash",
					"-c",
					fmt.Sprintf(
						`add-apt-repository "deb [arch=%s] https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-%s main")`,
						system.System.Arch,
						system.System.CodeName,
					),
				),
			); err != nil {
				return err
			}
		} else {
			if err := pkg.ExecCmd(
				exec.Command(
					"/bin/bash",
					"-c",
					fmt.Sprintf(
						`add-apt-repository "deb [arch=%s] https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main"`,
						system.System.Arch,
					),
				),
			); err != nil {
				return err
			}
		}
		if err := pkg.ExecCmd(exec.Command("apt", "update")); err != nil {
			return err
		}
		k8sVersion := strings.Replace(withKubernetesVersion, "v", "", -1)
		if err := pkg.ExecCmd(
			exec.Command(
				"apt",
				"install",
				"-y",
				"kubelet="+k8sVersion+"-00",
				"kubeadm="+k8sVersion+"-00",
				"kubectl="+k8sVersion+"-00",
				"kubernetes-cni",
			),
		); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("systemctl", "daemon-reload")); err != nil {
			return err
		}
		if err := pkg.ExecCmd(exec.Command("systemctl", "enable", "kubelet", "--now")); err != nil {
			return err
		}
		_ = pkg.ExecCmd(exec.Command("systemctl", "status", "kubelet"))
		break
	}

	// 加载容器镜像
	return loadImage(containerWithDocker)
}
