package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/apprenda/kismatic/pkg/data"
	"github.com/apprenda/kismatic/pkg/install"
	"github.com/spf13/cobra"
)

type volumeListOptions struct {
	outputFormat string
}

// NewCmdVolumeList returns the command for listgin storage volumes
func NewCmdVolumeList(out io.Writer, planFile *string) *cobra.Command {
	opts := volumeListOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list storage volumes to the Kubernetes cluster",
		Long: `List storage volumes to the Kubernetes cluster.

This function requires a target cluster that has storage nodes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doVolumeList(out, opts, *planFile, args)
		},
	}

	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "simple", `output format (options "simple"|"json")`)
	return cmd
}

func doVolumeList(out io.Writer, opts volumeListOptions, planFile string, args []string) error {
	// verify command
	if opts.outputFormat != "simple" && opts.outputFormat != "json" {
		return fmt.Errorf("output format %q is not supported", opts.outputFormat)
	}

	// Setup ansible
	planner := &install.FilePlanner{File: planFile}
	if !planner.PlanExists() {
		return fmt.Errorf("plan file not found at %q", planFile)
	}

	plan, err := planner.Read()
	if err != nil {
		return fmt.Errorf("error reading plan file: %v", err)
	}

	// find storage node
	clientStorage, err := plan.GetSSHClient("storage")
	if err != nil {
		return err
	}
	glusterClient := data.RemoteGlusterCLI{SSHClient: clientStorage}

	// find master node
	clientMaster, err := plan.GetSSHClient("master")
	if err != nil {
		return err
	}
	kubernetesClient := data.RemoteKubectl{SSHClient: clientMaster}

	resp, err := buildResponse(glusterClient, kubernetesClient)
	if err != nil {
		return err
	}
	if resp == nil {
		fmt.Fprintln(out, "No volumes were found on the cluster. You may use `kismatic volume add` to create new volumes.")
		return nil
	}

	return print(out, resp, opts.outputFormat)
}

func buildResponse(glusterClient data.GlusterClient, kubernetesClient data.KubernetesClient) (*ListResponse, error) {
	// get gluster volume data
	glusterVolumeInfo, err := glusterClient.ListVolumes()
	if err != nil {
		return nil, err
	}
	if glusterVolumeInfo == nil {
		return nil, nil
	}
	// get persistent volumes data
	pvs, err := kubernetesClient.ListPersistentVolumes()
	if err != nil {
		return nil, err
	}
	// get pods data
	pods, err := kubernetesClient.ListPods()
	if err != nil {
		return nil, err
	}

	// build a map of pods that have PersistentVolumeClaim
	podsMap := make(map[string][]Pod)
	// iterate through all the pods
	// since the api doesnt have a pv -> pod data, need to search through all the pods
	// this will get PV -> PVC - > pod(s) -> container(s)
	if pods != nil { // no pods running
		for _, pod := range pods.Items {
			for _, v := range pod.Spec.Volumes {
				if v.PersistentVolumeClaim != nil {
					var containers []Container
					for _, container := range pod.Spec.Containers {
						for _, volumeMount := range container.VolumeMounts {
							if volumeMount.Name == v.Name {
								containers = append(containers, Container{Name: container.Name, MountName: volumeMount.Name, MountPath: volumeMount.MountPath})
							}
						}
					}
					// pods that have the same PVC are in one list
					key := strings.Join([]string{pod.Namespace, v.PersistentVolumeClaim.ClaimName}, ":")
					p := Pod{Namespace: pod.Namespace, Name: pod.Name, Containers: containers}
					podsMap[key] = append(podsMap[key], p)
				}
			}
		}
	}

	// iterate through PVs once and build a map
	pvsMap := make(map[string]data.PersistentVolume)
	if pvs != nil {
		for _, pv := range pvs.Items {
			pvsMap[pv.Name] = pv
		}
	}

	// build response object
	resp := ListResponse{}
	// loop through all the gluster volumes
	for _, gv := range glusterVolumeInfo.VolumeInfo.Volumes.Volume {
		v := Volume{
			Name:              gv.Name,
			DistributionCount: gv.BrickCount / gv.ReplicaCount, //gv.DistCount doesn't actually return the correct number when ReplicaCount > 1
			ReplicaCount:      gv.ReplicaCount,
			Capacity:          "Unknown",
			Available:         "Unknown",
			Status:            "Unknown",
		}

		if gv.BrickCount > 0 {
			v.Bricks = make([]Brick, gv.BrickCount)
			for n, gbrick := range gv.Bricks.Brick {
				brickArr := strings.Split(gbrick.Text, ":")
				v.Bricks[n] = Brick{Host: brickArr[0], Path: brickArr[1]}
			}
		}
		// get gluster volume quota
		glusterVolumeQuota, err := glusterClient.GetQuota(gv.Name)
		if err != nil {
			return nil, err
		}
		if glusterVolumeQuota != nil && glusterVolumeQuota.VolumeQuota != nil && glusterVolumeQuota.VolumeQuota.Limit != nil {
			v.Capacity = HumanFormat(glusterVolumeQuota.VolumeQuota.Limit.HardLimit)
		}
		if glusterVolumeQuota != nil && glusterVolumeQuota.VolumeQuota != nil && glusterVolumeQuota.VolumeQuota.Limit != nil {
			v.Available = HumanFormat(glusterVolumeQuota.VolumeQuota.Limit.AvailSpace)
		}
		// it is possible that all PVs were delete in kubernetes
		// set status of gluster volume to "Unknown"
		foundPVInfo, ok := pvsMap[gv.Name]
		// this PV does not exist, maybe it was deleted?
		// set status of gluster volume to "Unknown"
		if ok {
			if class, ok := foundPVInfo.ObjectMeta.Annotations["volume.beta.kubernetes.io/storage-class"]; ok {
				v.StorageClass = class
			}
			v.Labels = foundPVInfo.Labels
			v.Status = string(foundPVInfo.Status.Phase)
			if foundPVInfo.Spec.ClaimRef != nil {
				// populate claim info
				v.Claim = &Claim{Namespace: foundPVInfo.Spec.ClaimRef.Namespace, Name: foundPVInfo.Spec.ClaimRef.Name}
				// populate pod info
				key := strings.Join([]string{foundPVInfo.Spec.ClaimRef.Namespace, foundPVInfo.Spec.ClaimRef.Name}, ":")
				if pod, ok := podsMap[key]; ok && pod != nil {
					v.Pods = pod
				}
			}
		}

		resp.Volumes = append(resp.Volumes, v)
	}
	// return nil if there are no volumes
	if len(resp.Volumes) == 0 {
		return nil, nil
	}
	return &resp, nil
}

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
func print(out io.Writer, resp *ListResponse, format string) error {
	if format == "simple" {
		separator := ""
		w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
		for _, v := range resp.Volumes {
			fmt.Fprint(w, separator)
			fmt.Fprintf(w, "Name:\t%s\t\n", v.Name)
			fmt.Fprintf(w, "StorageClass:\t%s\t\n", v.StorageClass)
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
		fmt.Fprintln(out, string(prettyResp))
	}

	return nil
}
