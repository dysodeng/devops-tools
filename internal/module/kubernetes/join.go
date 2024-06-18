package kubernetes

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
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
	var err error
	if withMaster {
		err = joinKubernetesControlPlaneNode()
	} else {
		err = joinKubernetesWorkerNode()
	}
	return err
}

// joinKubernetesControlPlaneNode 添加控制面节点
func joinKubernetesControlPlaneNode() error {
	// 生成新证书
	certCmd := exec.Command("kubeadm", "init", "phase", "upload-certs", "--upload-certs")
	certOut, err := certCmd.Output()
	if err != nil {
		return err
	}
	certOutLineList := strings.Split(strings.TrimSpace(string(certOut)), "\n")
	var certKey string
	for _, certOutLine := range certOutLineList {
		if ok, err := regexp.MatchString("^[a-zA-Z0-9]+$", certOutLine); err != nil {
			return err
		} else {
			if ok {
				certKey = certOutLine
				break
			}
		}
	}
	if certKey == "" {
		return errors.New("证书生成失败")
	}

	command, err := generateKubernetesJoinNodeCommand()
	if err != nil {
		return err
	}

	fmt.Printf("%s --control-plane --certificate-key %s\n", command, certKey)

	return nil
}

// joinKubernetesWorkerNode 添加工作节点
func joinKubernetesWorkerNode() error {
	command, err := generateKubernetesJoinNodeCommand()
	if err != nil {
		return err
	}
	log.Println(command)
	return nil
}

func generateKubernetesJoinNodeCommand() (string, error) {
	tokenCmd := exec.Command("kubeadm", "token", "create")
	tokenOut, err := tokenCmd.Output()
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(tokenOut))

	certCmd := exec.Command(
		"/bin/bash",
		"-c",
		`openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | openssl rsa -pubin -outform der 2>/dev/null | openssl dgst -sha256 -hex | sed 's/^.* //'`,
	)
	certOut, err := certCmd.Output()
	if err != nil {
		return "", err
	}
	certKey := strings.TrimSpace(string(certOut))

	serverAddr := k8sServerAddr()

	return fmt.Sprintf(`kubeadm join %s:6443 --token %s \
        --discovery-token-ca-cert-hash sha256:%s`, serverAddr, token, certKey), nil
}
