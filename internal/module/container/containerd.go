package container

import (
	"os/exec"
	"strings"

	"github.com/dysodeng/devops-tools/internal/pkg"
)

// installContainerd 安装containerd
func installContainerd(linuxDistro, arch string) error {
	var err error
	switch linuxDistro {
	case "CentOS":
		// 关闭防火墙
		if err = pkg.ExecCmd(exec.Command("systemctl", "stop", "firewalld.service")); err != nil {
			return err
		}
		if err = pkg.ExecCmd(exec.Command("systemctl", "disable", "firewalld.service")); err != nil {
			return err
		}

		// 关闭selinux
		if err = pkg.ExecCmd(exec.Command("setenforce", "0")); err != nil {
			return err
		}
		if err = pkg.ExecCmd(exec.Command("sed", "-i", "s/^SELINUX=enforcing$/SELINUX=permissive/", "/etc/selinux/config")); err != nil {
			return err
		}

		// 安装containerd
		if err = pkg.ExecCmd(exec.Command("yum", "install", "-y", "yum-utils", "device-mapper-persistent-data", "lvm2")); err != nil {
			return err
		}
		if err = pkg.ExecCmd(exec.Command("yum-config-manager", "--add-repo", "https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo")); err != nil {
			return err
		}
		if err = pkg.ExecCmd(exec.Command("yum", "install", "-y", "containerd.io", "runc")); err != nil {
			return err
		}
		if err = pkg.ExecCmd(exec.Command("systemctl", "stop", "containerd.service")); err != nil {
			return err
		}

		// 配置containerd
		if err = containerdConfig(); err != nil {
			return err
		}

		if err = crictlConfig(); err != nil {
			return err
		}

		// 启动containerd服务
		return pkg.ExecCmd(exec.Command("systemctl", "enable", "--now", "containerd.service"))

	case "Ubuntu":
		fallthrough

	case "Debian":

		// 关闭防火墙
		_ = pkg.ExecCmd(exec.Command("systemctl", "disable", "ufw", "--now"))

		if err = pkg.ExecCmd(exec.Command("apt-get", "update")); err != nil {
			return err
		}

		// 安装ubuntu发行版最新containerd
		searchCmd := exec.Command("apt-cache", "madison", "containerd")
		output, e := searchCmd.Output()
		if e != nil {
			return e
		}
		searchList := strings.Split(string(output), "\n")
		var containerdVersion string
		if len(searchList) > 0 {
			versionList := strings.Split(searchList[0], "|")
			if len(versionList) >= 3 {
				containerdVersion = "containerd=" + strings.TrimSpace(versionList[1])
			}
		}
		if err = pkg.ExecCmd(exec.Command("apt", "install", "-y", containerdVersion)); err != nil {
			return err
		}

		// 配置containerd
		if err = containerdConfig(); err != nil {
			return err
		}

		if err = crictlConfig(); err != nil {
			return err
		}

		// 启动containerd服务
		return pkg.ExecCmd(exec.Command("systemctl", "enable", "--now", "containerd"))
	}

	return nil
}
