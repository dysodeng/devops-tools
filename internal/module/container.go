package module

import (
	"fmt"
	"github.com/dysodeng/devops-tools/internal/pkg"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"strings"
)

var ContainerCmd = &cobra.Command{
	Use:   "container",
	Short: "容器运行时配置",
	Long:  "容器运行时配置",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var containerWithDocker bool // 使用Docker，否则使用containerd

// installContainerCmd 安装容器运行时
var installContainerCmd = &cobra.Command{
	Use:   "install",
	Short: "安装容器运行时，默认安装containerd",
	Long:  "安装容器运行时，默认安装containerd",
	Run: func(cmd *cobra.Command, args []string) {
		// 获取当前操作系统
		if system.OS != "linux" {
			log.Println("操作系统不是Linux")
			os.Exit(1)
		}

		var err error
		if containerWithDocker {
			err = installDocker(system.LinuxDistro, system.Arch)
		} else {
			err = installContainerd(system.LinuxDistro, system.Arch)
		}
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

func initContainerCmd() {
	installContainerCmd.Flags().BoolVarP(&containerWithDocker, "with-docker", "", false, "安装Docker")
	ContainerCmd.AddCommand(installContainerCmd)
}

func installDocker(linuxDistro, arch string) error {
	return nil
}

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
		if err = pkg.ExecCmd(exec.Command("yum-config-manager", "--add-repo", "https://download.docker.com/linux/centos/docker-ce.repo")); err != nil {
			return err
		}
		if err = pkg.ExecCmd(exec.Command("yum", "install", "-y", "containerd.io", "runc")); err != nil {
			return err
		}
		if err = pkg.ExecCmd(exec.Command("systemctl", "stop", "containerd.service")); err != nil {
			return err
		}

		// 配置containerd
		_ = os.Mkdir("/etc/containerd", os.ModeDir)
		if err = os.Rename("/etc/containerd/config.toml", "/etc/containerd/config.toml.bak"); err != nil {
			return err
		}
		_ = pkg.ExecCmd(exec.Command("touch", "/etc/containerd/config.toml"))
		configFilePath := "/etc/containerd/config.toml"
		configFile, e := os.Create(configFilePath)
		if e != nil {
			return e
		}
		configCmd := exec.Command("containerd", "config", "default")
		configCmd.Stdout = configFile
		if err = configCmd.Run(); err != nil {
			return err
		}
		if err = pkg.ExecCmd(exec.Command("sed", "-i", "s#registry.k8s.io/pause:3.6#registry.aliyuncs.com/google_containers/pause:3.9#g", "/etc/containerd/config.toml")); err != nil {
			return err
		}
		if err = pkg.ExecCmd(exec.Command("sed", "-i", "s#SystemdCgroup = false#SystemdCgroup = true#g", "/etc/containerd/config.toml")); err != nil {
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
		_ = os.Mkdir("/etc/containerd", os.ModeDir)
		_ = os.Rename("/etc/containerd/config.toml", "/etc/containerd/config.toml.bak")
		_ = pkg.ExecCmd(exec.Command("touch", "/etc/containerd/config.toml"))
		configFilePath := "/etc/containerd/config.toml"
		configFile, e := os.Create(configFilePath)
		if e != nil {
			return e
		}
		configCmd := exec.Command("containerd", "config", "default")
		configCmd.Stdout = configFile
		if err = configCmd.Run(); err != nil {
			return err
		}
		if err = pkg.ExecCmd(exec.Command("sed", "-i", "s#registry.k8s.io/pause:3.8#registry.aliyuncs.com/google_containers/pause:3.9#g", "/etc/containerd/config.toml")); err != nil {
			return err
		}
		if err = pkg.ExecCmd(exec.Command("sed", "-i", "s#SystemdCgroup = false#SystemdCgroup = true#g", "/etc/containerd/config.toml")); err != nil {
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

func crictlConfig() error {
	configFile, e := os.OpenFile("/etc/crictl.yaml", os.O_CREATE|os.O_RDWR, 0644)
	if e != nil {
		return e
	}
	defer func() {
		_ = configFile.Close()
	}()

	if _, err := configFile.Write([]byte(`runtime-endpoint: unix:///run/containerd/containerd.sock
image-endpoint: unix:///run/containerd/containerd.sock
timeout: 10
debug: false`)); err != nil {
		return err
	}

	return nil
}
