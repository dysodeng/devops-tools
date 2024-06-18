package container

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dysodeng/devops-tools/internal/pkg"
)

const (
	ContainerdSockPath   = "/run/containerd/containerd.sock"
	ContainerdConfigPath = "/etc/containerd"
)

// containerdConfig containerd配置
func containerdConfig() error {
	_ = os.Mkdir(ContainerdConfigPath, os.ModeDir)
	_ = os.Rename(
		fmt.Sprintf("%s/config.toml", ContainerdConfigPath),
		fmt.Sprintf("%s/config.toml.bak", ContainerdConfigPath),
	)
	_ = pkg.ExecCmd(
		exec.Command("touch", fmt.Sprintf("%s/config.toml", ContainerdConfigPath)),
	)

	configFilePath := fmt.Sprintf("%s/config.toml", ContainerdConfigPath)
	configFile, e := os.Create(configFilePath)
	if e != nil {
		return e
	}
	configCmd := exec.Command("containerd", "config", "default")
	configCmd.Stdout = configFile
	if err := configCmd.Run(); err != nil {
		return err
	}

	if err := pkg.ExecCmd(
		exec.Command(
			"sed",
			"-i",
			"s#registry.k8s.io/pause:3.8#registry.aliyuncs.com/google_containers/pause:3.9#g",
			fmt.Sprintf("%s/config.toml", ContainerdConfigPath),
		),
	); err != nil {
		return err
	}
	if err := pkg.ExecCmd(
		exec.Command(
			"sed",
			"-i",
			"s#SystemdCgroup = false#SystemdCgroup = true#g",
			fmt.Sprintf("%s/config.toml", ContainerdConfigPath),
		),
	); err != nil {
		return err
	}

	// 指定数据目录
	if containerWithDataDirectory != "" {
		if err := pkg.ExecCmd(
			exec.Command(
				"sed",
				"-i",
				fmt.Sprintf("s#/var/lib/containerd#%s#g", containerWithDataDirectory),
				fmt.Sprintf("%s/config.toml", ContainerdConfigPath),
			),
		); err != nil {
			return err
		}
	}

	return nil
}

// crictlConfig 配置crictl
func crictlConfig() error {
	configFile, e := os.OpenFile("/etc/crictl.yaml", os.O_CREATE|os.O_RDWR, 0644)
	if e != nil {
		return e
	}
	defer func() {
		_ = configFile.Close()
	}()

	if _, err := configFile.Write([]byte(fmt.Sprintf(`runtime-endpoint: unix://%s
image-endpoint: unix://%s
timeout: 10
debug: false`, ContainerdSockPath, ContainerdSockPath))); err != nil {
		return err
	}

	return nil
}
