package kubernetes

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

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

// joinKubernetesNode 加入k8s节点
func joinKubernetesNode(withMaster bool) error {
	tokenCmd := exec.Command("kubeadm", "token", "create")
	tokenOut, err := tokenCmd.Output()
	if err != nil {
		return err
	}
	token := strings.TrimSpace(string(tokenOut))

	certCmd := exec.Command(
		"/bin/bash",
		"-c",
		`openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null | openssl dgst -sha256 -hex | sed 's/^.* //'`,
	)
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
