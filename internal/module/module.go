package module

func init() {
	system = systemInfo()
	initSystemCmd()
	initContainerCmd()
	initKubernetes()
}
