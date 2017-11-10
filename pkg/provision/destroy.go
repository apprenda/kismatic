package provision

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type DestroyOpts struct {
	ClusterName string
}

//Destroy destroys a provisioned cluster (using -force by default)
func Destroy(out io.Writer, opts *DestroyOpts) error {
	clusterPathFromWd := fmt.Sprintf("terraform/clusters/%s/", opts.ClusterName)
	os.Chdir(clusterPathFromWd)
	tfDestroy := exec.Command(terraform, "destroy", "-force")
	if stdoutStderr, err := tfDestroy.CombinedOutput(); err != nil {
		return fmt.Errorf("Error attempting to destroy: %s", stdoutStderr)
	}
	fmt.Fprintf(out, "Cluster destruction successful.\n")
	os.Chdir("../../../")
	//os.RemoveAll(clusterPathFromWd)
	return nil
}
