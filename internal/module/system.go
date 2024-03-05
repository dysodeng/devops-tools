package module

import (
	"errors"
	"fmt"
	"github.com/dysodeng/devops-tools/internal/pkg"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// System 操作系统信息
type System struct {
	OS                   string // 操作系统类型
	Arch                 string // 平台架构
	LinuxDistro          string // Linux发行版名称
	LinuxDistroVersion   string // Linux发行版(饮食版本)
	LinuxKernel          string // Linux内核版本
	LinuxKernelMasterNum int    // Linux内核主要版本
	CpuCores             int    // Cpu核心数
}

var SystemCmd = &cobra.Command{
	Use:   "system",
	Short: "操作系统配置",
	Long:  "操作系统配置",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "系统信息",
	Long:  "系统信息",
	Run: func(cmd *cobra.Command, args []string) {
		// 获取当前操作系统
		info := systemInfo()
		tablePrefix := "\t"
		if info.OS == "linux" {
			tablePrefix = "\t\t"
		}
		fmt.Println("---------- 系统信息 ----------")
		fmt.Printf("OS:%s%s\n", tablePrefix, info.OS)
		fmt.Printf("Arch:%s%s\n", tablePrefix, info.Arch)
		if info.OS == "linux" {
			fmt.Printf("Linux Dist:\t%s\n", info.LinuxDistroVersion)
			fmt.Printf("Linux Kernel:\t%s\n", info.LinuxKernel)
		}
		fmt.Printf("Cpus:%s%d\n", tablePrefix, info.CpuCores)
	},
}

var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "安装系统必要的工具",
	Long:  "安装系统必要的工具",
	Run: func(cmd *cobra.Command, args []string) {
		// 获取当前操作系统
		info := systemInfo()
		if info.OS != "linux" {
			fmt.Println("操作系统不是Linux")
			os.Exit(1)
		}

		err := toolInstall(info.LinuxDistro)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

var (
	initWithSource        string
	initWithDefaultSource bool
)

const (
	initWithCentOSDefaultSource string = "https://mirrors.aliyun.com/repo/Centos-7.repo"
	initWithUbuntuDefaultSource string = ""
	initWithDebianDefaultSource string = ""
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "系统初始化",
	Long:  "系统初始化",
	Run: func(cmd *cobra.Command, args []string) {
		// 获取当前操作系统
		info := systemInfo()
		if info.OS != "linux" {
			log.Println("操作系统不是Linux")
			os.Exit(1)
		}
		if info.Arch != "amd64" {
			log.Println("只支持x86_64的系统")
			os.Exit(1)
		}

		// 更换软件源
		err := changeSource(info.LinuxDistro, initWithDefaultSource, initWithSource)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// 升级Linux内核版本
		if info.LinuxKernelMasterNum < 4 {
			err = upgradeLinuxKernel(info.LinuxDistro)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
	},
}

func initSystemCmd() {
	initCmd.Flags().BoolVarP(&initWithDefaultSource, "default-source", "", false, "default-source")
	initCmd.Flags().StringVarP(&initWithSource, "source", "", "", "source")
	SystemCmd.AddCommand(
		infoCmd,
		toolCmd,
		initCmd,
	)
}

// systemInfo 获取操作系统信息
func systemInfo() System {
	osType := runtime.GOOS
	var linuxDistro, linuxDistroVersion, linuxKernel string
	var linuxKernelMasterNum int
	if osType == "linux" {
		// 获取linux发行版本
		distroCmd := exec.Command("lsb_release", "-d")
		distroOutput, err := distroCmd.Output()
		if err == nil {
			out := string(distroOutput)
			linuxDistroVersion = strings.TrimSpace(strings.Replace(out, "Description:", "", -1))
			switch true {
			case strings.Contains(linuxDistroVersion, "CentOS"):
				linuxDistro = "CentOS"
				break
			case strings.Contains(linuxDistroVersion, "Ubuntu"):
				linuxDistro = "Ubuntu"
				break
			case strings.Contains(linuxDistroVersion, "Debian"):
				linuxDistro = "Debian"
				break
			}
		}

		// 获取linux内核版本
		kernelCmd := exec.Command("uname", "-r")
		kernelOutput, err := kernelCmd.Output()
		if err == nil {
			linuxKernel = strings.TrimSpace(string(kernelOutput))
			kernel := strings.Split(linuxKernel, ".")
			kernelNum, err := strconv.ParseInt(kernel[0], 10, 64)
			if err == nil {
				linuxKernelMasterNum = int(kernelNum)
			}
		}
	}
	return System{
		OS:                   osType,
		Arch:                 runtime.GOARCH,
		LinuxDistro:          linuxDistro,
		LinuxDistroVersion:   linuxDistroVersion,
		LinuxKernel:          linuxKernel,
		LinuxKernelMasterNum: linuxKernelMasterNum,
		CpuCores:             runtime.NumCPU(),
	}
}

// toolInstall 工具安装
func toolInstall(linuxDistro string) error {
	var err error
	switch linuxDistro {
	case "CentOS":
		err = pkg.ExecCmd(exec.Command("yum", "install", "-y", "wget", "curl", "vim", "net-tools"))
		break
	case "Ubuntu":
		err = pkg.ExecCmd(exec.Command("apt", "install", "-y", "wget", "curl", "vim", "net-tools"))
		break
	case "Debian":
		err = pkg.ExecCmd(exec.Command("apt", "install", "-y", "wget", "curl", "vim", "net-tools"))
		break
	default:
		err = errors.New("不支持的Linux发行版")
	}
	return err
}

func cleanSystemSource(linuxDistro string) error {
	var err error

	switch linuxDistro {
	case "CentOS":
		err = pkg.ExecCmd(exec.Command("yum", "clean", "all"))
		if err != nil {
			return err
		}
		err = pkg.ExecCmd(exec.Command("yum", "makecache"))
		if err != nil {
			return err
		}
		break
	}

	return nil
}

// changeSource 更换软件源
func changeSource(linuxDistro string, isDefaultSource bool, customSource string) error {
	var descSource string
	if !isDefaultSource {
		descSource = customSource
	}
	if isDefaultSource {
		switch linuxDistro {
		case "CentOS":
			descSource = initWithCentOSDefaultSource
			break
		case "Ubuntu":
			descSource = initWithUbuntuDefaultSource
			break
		case "Debian":
			descSource = initWithDebianDefaultSource
			break
		}
	} else {
		descSource = customSource
	}

	var err error
	if descSource == "" {
		return nil
	}

	switch linuxDistro {
	case "CentOS":
		err = pkg.ExecCmd(exec.Command("mv", "/etc/yum.repos.d/CentOS-Base.repo", "/etc/yum.repos.d/CentOS-Base.repo.bak"))
		if err != nil {
			return err
		}

		err = pkg.ExecCmd(exec.Command("wget", "-O", "/etc/yum.repos.d/CentOS-Base.repo", descSource))
		if err != nil {
			return err
		}
		err = cleanSystemSource(linuxDistro)
		if err != nil {
			return err
		}
		err = pkg.ExecCmd(exec.Command("yum", "update", "-y"))
		if err != nil {
			return err
		}
		break

	case "Ubuntu":
		break

	case "Debian":
		break

	default:
		err = errors.New("不支持的Linux发行版")
	}

	return err
}

// upgradeLinuxKernel 升级Linux内核版本
func upgradeLinuxKernel(linuxDistro string) error {
	var err error
	switch linuxDistro {
	case "CentOS":

		// 内核源
		err = elrepo()
		if err != nil {
			return err
		}

		err = cleanSystemSource(linuxDistro)
		if err != nil {
			return err
		}

		err = pkg.ExecCmd(exec.Command("yum", "install", "-y", "kernel-lt-5.4.262"))
		if err != nil {
			return err
		}
		err = pkg.ExecCmd(exec.Command("yum", "install", "-y", "kernel-lt-devel-5.4.262"))
		if err != nil {
			return err
		}
		_ = pkg.ExecCmd(exec.Command("grub2-set-default", "0"))
		fmt.Println("内核已更新，重启后生效")
		break
	}
	return err
}

func elrepo() error {
	repoFile, err := os.OpenFile("/etc/yum.repos.d/elrepo.repo", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = repoFile.Close()
	}()

	_, err = repoFile.Write([]byte(`[elrepo]
name=elrepo
baseurl=https://mirrors.aliyun.com/elrepo/archive/kernel/el7/x86_64
gpgcheck=0
enabled=1`))
	return err
}
