package kubernetes

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"

	"github.com/dysodeng/devops-tools/internal/pkg"
	"github.com/spf13/cobra"
)

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

// initKubernetesCluster 初始化k8s集群
func initKubernetesCluster(k8sVersion string) error {

	serverAddr := k8sServerAddr()
	var err error

	// 初始化k8s集群
	fmt.Println("\n初始化Kubernetes集群...")
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
	if err = pkg.ExecCmd(
		exec.Command(
			"chown",
			fmt.Sprintf("%s:%s", currentUser.Uid, currentUser.Gid), homePath+"/.kube/config",
		),
	); err != nil {
		return err
	}

	// 初始化集群网络
	if err = pkg.ExecCmd(exec.Command("kubectl", "apply", "-f", "./config/calico.yaml")); err != nil {
		return err
	}
	if err = pkg.ExecCmd(exec.Command("kubectl", "get", "nodes")); err != nil {
		return err
	}

	return nil
}
