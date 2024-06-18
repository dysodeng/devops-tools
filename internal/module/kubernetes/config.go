package kubernetes

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dysodeng/devops-tools/internal/module/system"
)

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

	arch := system.ArchMap[system.System.Arch]

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
