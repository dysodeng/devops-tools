package module

import (
	"fmt"
	"github.com/dysodeng/devops-tools/internal/pkg"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

// KubernetesCmd k8s配置命令
var KubernetesCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Kubernetes配置",
	Long:  "Kubernetes配置",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

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

// initKubernetesClusterCmd 初始化k8s集群
var initKubernetesClusterCmd = &cobra.Command{
	Use:   "init-cluster",
	Short: "初始化Kubernetes集群",
	Long:  "初始化Kubernetes集群",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initKubernetesCluster(withKubernetesVersion); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

// joinMasterNode 加入master节点
var joinMasterNode bool

// joinKubernetesNodeCmd Kubernetes加入节点命令
var joinKubernetesNodeCmd = &cobra.Command{
	Use:   "join-node",
	Short: "Kubernetes加入节点",
	Long:  "Kubernetes加入节点",
	Run: func(cmd *cobra.Command, args []string) {
		if err := joinKubernetesNode(joinMasterNode); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

func initKubernetes() {
	installKubernetesCmd.Flags().StringVarP(&withKubernetesVersion, "with-version", "", "v1.27.6", "指定Kubernetes版本")
	initKubernetesClusterCmd.Flags().StringVarP(&withKubernetesVersion, "with-version", "", "v1.27.6", "指定Kubernetes版本")
	joinKubernetesNodeCmd.Flags().BoolVarP(&joinMasterNode, "master", "", false, "加入master节点")
	KubernetesCmd.AddCommand(installKubernetesCmd, initKubernetesClusterCmd, joinKubernetesNodeCmd)
}

// installKubernetes 安装k8s组件
func installKubernetes() error {
	switch system.LinuxDistro {
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
		if err := pkg.ExecCmd(exec.Command("yum", "install", "-y", "kubelet-"+k8sVersion+"-0", "kubeadm-"+k8sVersion+"-0", "kubectl-"+k8sVersion+"-0", "kubernetes-cni", "--disableexcludes=kubernetes", "--nogpgcheck")); err != nil {
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
		if err := pkg.ExecCmd(exec.Command("/bin/bash", "-c", `curl -fsSL https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add -`)); err != nil {
			return err
		}
		if pkg.CheckNetworkFileExists("https://mirrors.aliyun.com/kubernetes/apt/kubernetes-" + system.CodeName + "/Release") {
			if err := pkg.ExecCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf(`add-apt-repository "deb [arch=%s] https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-%s main")`, system.Arch, system.CodeName))); err != nil {
				return err
			}
		} else {
			if err := pkg.ExecCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf(`add-apt-repository "deb [arch=%s] https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main"`, system.Arch))); err != nil {
				return err
			}
		}
		if err := pkg.ExecCmd(exec.Command("apt", "update")); err != nil {
			return err
		}
		k8sVersion := strings.Replace(withKubernetesVersion, "v", "", -1)
		if err := pkg.ExecCmd(exec.Command("apt", "install", "-y", "kubelet="+k8sVersion+"-00", "kubeadm="+k8sVersion+"-00", "kubectl="+k8sVersion+"-00", "kubernetes-cni")); err != nil {
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
		if err := pkg.ExecCmd(exec.Command("apt", "install", "-y", "software-properties-common", "dirmngr", "lsb-release", "ca-certificates")); err != nil {
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
		if err := pkg.ExecCmd(exec.Command("/bin/bash", "-c", `curl -fsSL https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add -`)); err != nil {
			return err
		}
		if pkg.CheckNetworkFileExists("https://mirrors.aliyun.com/kubernetes/apt/kubernetes-" + system.CodeName + "/Release") {
			if err := pkg.ExecCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf(`add-apt-repository "deb [arch=%s] https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-%s main")`, system.Arch, system.CodeName))); err != nil {
				return err
			}
		} else {
			if err := pkg.ExecCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf(`add-apt-repository "deb [arch=%s] https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main"`, system.Arch))); err != nil {
				return err
			}
		}
		if err := pkg.ExecCmd(exec.Command("apt", "update")); err != nil {
			return err
		}
		k8sVersion := strings.Replace(withKubernetesVersion, "v", "", -1)
		if err := pkg.ExecCmd(exec.Command("apt", "install", "-y", "kubelet="+k8sVersion+"-00", "kubeadm="+k8sVersion+"-00", "kubectl="+k8sVersion+"-00", "kubernetes-cni")); err != nil {
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
	return nil
}

// initKubernetesCluster 初始化k8s集群
func initKubernetesCluster(k8sVersion string) error {

	serverAddr := k8sServerAddr()
	var err error

	// 初始化k8s集群
	fmt.Println("初始化Kubernetes集群...")
	if err = pkg.ExecCmd(exec.Command("/bin/bash", "-c", fmt.Sprintf(`kubeadm init \
    --image-repository=registry.aliyuncs.com/google_containers \
    --apiserver-advertise-address=%s \
    --kubernetes-version=%s \
    --service-cidr=10.96.0.0/16 \
    --pod-network-cidr=10.244.0.0/16`, serverAddr, k8sVersion))); err != nil {
		return err
	}

	// 初始化配置
	homePath := os.Getenv("HOME")
	currentUser, _ := user.Current()
	if err = pkg.ExecCmd(exec.Command("mkdir", "-p", homePath+"/.kube")); err != nil {
		return err
	}
	if err = pkg.ExecCmd(exec.Command("cp", "-i", "/etc/kubernetes/admin.conf", homePath+"/.kube/config")); err != nil {
		return err
	}
	if err = pkg.ExecCmd(exec.Command("chown", fmt.Sprintf("%s:%s", currentUser.Uid, currentUser.Gid), homePath+"/.kube/config")); err != nil {
		return err
	}

	// 初始化集群网络
	fmt.Println("初始化Kubernetes集群网络...")
	if err = pkg.ExecCmd(exec.Command("kubectl", "apply", "-f", "./config/calico.yaml")); err != nil {
		return err
	}
	if err = pkg.ExecCmd(exec.Command("kubectl", "get", "nodes")); err != nil {
		return err
	}

	return nil
}

// joinKubernetesNode 加入k8s节点
func joinKubernetesNode(withMaster bool) error {
	tokenCmd := exec.Command("kubeadm", "token", "create")
	tokenOut, err := tokenCmd.Output()
	if err != nil {
		return err
	}
	token := strings.TrimSpace(string(tokenOut))

	certCmd := exec.Command("/bin/bash", "-c", `openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null | openssl dgst -sha256 -hex | sed 's/^.* //'`)
	certOut, err := certCmd.Output()
	if err != nil {
		return err
	}
	certKey := strings.TrimSpace(string(certOut))

	serverAddr := k8sServerAddr()

	var masterTag string
	if withMaster {
		masterTag = " --control-plane"
	}

	fmt.Printf(`kubeadm join %s:6443 --token %s \
        --discovery-token-ca-cert-hash sha256:%s%s
`, serverAddr, token, certKey, masterTag)

	return nil
}

func k8sSysctlConfig() error {
	configFile, e := os.OpenFile("/etc/sysctl.d/k8s.conf", os.O_CREATE|os.O_RDWR, 0644)
	if e != nil {
		return e
	}
	defer func() {
		_ = configFile.Close()
	}()

	if _, err := configFile.Write([]byte(`net.bridge.bridge-nf-call-iptables=1
net.bridge.bridge-nf-call-ip6tables=1
net.ipv4.ip_forward=1
vm.swappiness=0`)); err != nil {
		return err
	}

	return nil
}

func k8sModuleLoadConfig() error {
	configFile, e := os.OpenFile("/etc/modules-load.d/k8s.conf", os.O_CREATE|os.O_RDWR, 0644)
	if e != nil {
		return e
	}
	defer func() {
		_ = configFile.Close()
	}()

	if _, err := configFile.Write([]byte(`overlay
br_netfilter
ip_tables
iptable_filter`)); err != nil {
		return err
	}

	return nil
}

func k8sRepoCentosConfig() error {
	configFile, e := os.OpenFile("/etc/yum.repos.d/kubernetes.repo", os.O_CREATE|os.O_RDWR, 0644)
	if e != nil {
		return e
	}
	defer func() {
		_ = configFile.Close()
	}()

	arch := ArchMap[system.Arch]

	if _, err := configFile.Write([]byte(fmt.Sprintf(`[kubernetes]
name=Kubernetes
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-%s/
enabled=1
gpgcheck=0
repo_gpgcheck=0
gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg`, arch))); err != nil {
		return err
	}

	return nil
}

func k8sServerAddr() string {
	cmd := exec.Command("/bin/bash", "-c", `ifconfig eth0 | grep "inet" | cut -d ':' -f 2 | cut -d '' -f 1 | awk '{print $2}'`)
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}
