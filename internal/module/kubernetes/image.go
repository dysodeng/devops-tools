package kubernetes

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/containerd/containerd"
	"github.com/dysodeng/devops-tools/internal/module/container"
	"github.com/spf13/cobra"
)

// containerWithDocker 使用Docker，否则使用containerd
var containerWithDocker bool

var loadImageCmd = &cobra.Command{
	Use:   "load-image",
	Short: "加载容器镜像",
	Long:  "加载容器镜像",
	Run: func(cmd *cobra.Command, args []string) {
		if err := loadImage(containerWithDocker); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

// loadImages 加载容器镜像
func loadImage(withDocker bool) error {
	log.Println("正在加载容器镜像...")
	var err error
	if withDocker {

	} else {

		client, clientErr := containerd.New(container.ContainerdSockPath, containerd.WithDefaultNamespace("k8s.io"))
		if clientErr != nil {
			return clientErr
		}

		err = filepath.Walk("./image", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Ext(path) == ".tar" {
				imageFile, fErr := os.OpenFile(path, os.O_RDONLY, 0)
				if fErr != nil {
					return fErr
				}
				list, loadError := client.Import(context.Background(), imageFile)
				if loadError != nil {
					return loadError
				}
				for _, image := range list {
					log.Printf("loaded image %s", image.Name)
				}
			}

			return nil
		})
	}

	if err != nil {
		return err
	}

	return nil
}
