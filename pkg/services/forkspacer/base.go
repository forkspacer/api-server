package forkspacer

const (
	BaseLabel = "forkspacer"
)

var Labels = struct {
	WorkspaceKubeconfigSecret string
}{
	WorkspaceKubeconfigSecret: "workspace-kubeconfig-secret",
}

type ResourceReference struct {
	Name      string
	Namespace string
}
