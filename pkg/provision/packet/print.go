package packet

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/apprenda/kismatic/pkg/provision/plan"
)

func printNodes(out io.Writer, nodes []plan.Node) {

	tw := tabwriter.NewWriter(out, 10, 4, 3, ' ', 0)
	fmt.Fprint(tw, "HOSTNAME\tPUBLIC IP\tPRIVATE IP\n")
	for _, n := range nodes {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", n.Host, n.PublicIPv4, n.PrivateIPv4)
	}
	tw.Flush()
}
