package volume

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
)

const (
	_          = iota // ignore first value by assigning to blank identifier
	KB float64 = 1 << (10 * iota)
	MB
	GB
	TB
)

// HumanFormat converts bytes to human readable KB,MB,GB,TB formats
func HumanFormat(bytes float64) string {
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2fTB", bytes/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2fGB", bytes/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2fMB", bytes/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2fKB", bytes/KB)
	}
	return fmt.Sprintf("%.2fB", bytes)
}

// Print prints the volume list response
func Print(out io.Writer, resp *ListResponse, format string) error {
	if format == "simple" {
		separator := ""
		w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
		for _, v := range resp.Volumes {
			fmt.Fprint(w, separator)
			fmt.Fprintf(w, "Name:\t%s\t\n", v.Name)
			if len(v.Labels) > 0 {
				fmt.Fprintf(w, "Labels:\t\t\n")
				for k, v := range v.Labels {
					fmt.Fprintf(w, "  %s:\t%s\t\n", k, v)
				}
			}
			fmt.Fprintf(w, "Capacity:\t%s\t\n", v.Capacity)
			fmt.Fprintf(w, "Available:\t%s\t\n", v.Available)
			fmt.Fprintf(w, "Replica:\t%d\t\n", v.ReplicaCount)
			fmt.Fprintf(w, "Distribution:\t%d\t\n", v.DistributionCount)
			fmt.Fprintf(w, "Bricks:\t%s\t\n", VolumeBrickToString(v.Bricks))
			fmt.Fprintf(w, "Status:\t%s\t\n", v.Status)
			fmt.Fprintf(w, "Claim:\t%s\t\n", v.Claim.Readable())
			fmt.Fprintf(w, "Pods:\t\t\n")
			for _, pod := range v.Pods {
				fmt.Fprintf(w, "  %s\t\n", pod.Readable())
				fmt.Fprintf(w, "    Containers:\t\t\n")
				for _, container := range pod.Containers {
					fmt.Fprintf(w, "      %s\t\t\n", container.Name)
					fmt.Fprintf(w, "        MountName:\t%s\t\n", container.MountName)
					fmt.Fprintf(w, "        MountPath:\t%s\t\n", container.MountPath)
				}
			}
			separator = "\n\n"
		}
		w.Flush()
	} else if format == "json" {
		// pretty prtin JSON
		prettyResp, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			return fmt.Errorf("marshal error: %v", err)
		}
		fmt.Println(string(prettyResp))
	}

	return nil
}
