package module

import (
	"github.com/dysodeng/devops-tools/internal/module/container"
	"github.com/dysodeng/devops-tools/internal/module/kubernetes"
	"github.com/dysodeng/devops-tools/internal/module/system"
)

func init() {
	system.InitSystemCmd()
	container.InitContainerCmd()
	kubernetes.InitKubernetesCmd()
}
