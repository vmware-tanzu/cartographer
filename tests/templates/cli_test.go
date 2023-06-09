package templates

import (
	"testing"

	cartotesting "github.com/vmware-tanzu/cartographer/pkg/testing"
)

func TestCLIExample(t *testing.T) {
	directories := []string{"kpack", "deliverable", "deployment"}

	for _, directory := range directories {
		err := cartotesting.CliTest(directory)
		if err != nil {
			t.Fatal("cli test failed,", "directory:", directory, "err:", err)
		}
	}
}
