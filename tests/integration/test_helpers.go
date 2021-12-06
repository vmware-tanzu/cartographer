package integration

import (
	"io"
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const DebugControlPlane = false

func CreateTestEnv(workingDir string, out io.Writer) *envtest.Environment {
	testEnv := &envtest.Environment{
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "..", "config", "webhook")},
		},
		CRDDirectoryPaths: []string{
			filepath.Join(workingDir, "..", "..", "..", "config", "crd", "bases"),
			filepath.Join(workingDir, "..", "..", "resources", "crds"),
		},
		AttachControlPlaneOutput: DebugControlPlane, // Set to true for great debug logging
	}

	if DebugControlPlane {
		apiServer := testEnv.ControlPlane.GetAPIServer()
		apiServer.Out = out
		apiServer.Err = out
		apiServer.Configure().
			Append("audit-policy-file", filepath.Join(workingDir, "..", "policy.yaml")).
			Append("audit-log-path", "-")
	}

	return testEnv
}
