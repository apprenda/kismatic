package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/apprenda/kismatic/pkg/data"
	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/volume"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/api/v1"
)

type volumeListOptions struct {
	outputFormat string
	namespace    string
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

	//cmd.Flags().StringVarP(&opts.namespace, "namespace", "ns", "all", `limit output to a single namespace`)
	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "simple", `output format (options "simple"|"json")`)
	return cmd
}

func doVolumeList(out io.Writer, opts volumeListOptions, planFile string, args []string) error {
	// Setup ansible
	planner := &install.FilePlanner{File: planFile}
	if !planner.PlanExists() {
		return fmt.Errorf("plan file not found at %q", planFile)
	}

	// verify command
	if opts.outputFormat != "simple" && opts.outputFormat != "json" {
		return fmt.Errorf("output format %q is not supported", opts.outputFormat)
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
	glusterGetter := data.GlusterCLIGetter{SSHClient: clientStorage}

	// find master node
	clientMaster, err := plan.GetSSHClient("master")
	if err != nil {
		return err
	}
	pvGetter := data.PlanPVGetter{SSHClient: clientMaster}
	podGetter := data.PlanPodGetter{SSHClient: clientMaster}

	resp, err := buildResponse(glusterGetter, pvGetter, podGetter)
	if err != nil {
		return err
	}
	if resp == nil {
		fmt.Fprintln(out, "No resources found.")
		return nil
	}

	return volume.Print(out, resp, opts.outputFormat)
}

func buildResponse(glusterGetter data.GlusterInfoGetter, pvGetter data.PVGetter, podGetter data.PodGetter) (*volume.ListResponse, error) {
	// get gluster volume data
	glusterVolumeInfo, err := glusterGetter.GetVolumes()
	if err != nil {
		return nil, err
	}
	if glusterVolumeInfo == nil {
		return nil, nil
	}
	// get persistent volumes data
	pvs, err := pvGetter.Get()
	if err != nil {
		return nil, err
	}
	// get pods data
	pods, err := podGetter.Get()
	if err != nil {
		return nil, err
	}

	// build a map of pods that have PersistentVolumeClaim
	podsMap := make(map[string][]volume.Pod)
	// iterate through all the pods
	// since the api doesnt have a pv -> pod data, need to search through all the pods
	// this will get PV -> PVC - > pod(s) -> container(s)
	if pods != nil { // possible for no pods to have claimed a PV
		for _, pod := range pods.Items {
			if len(pod.Spec.Volumes) > 0 {
				for _, v := range pod.Spec.Volumes {
					if v.PersistentVolumeClaim != nil {
						var containers []volume.Container
						for _, container := range pod.Spec.Containers {
							for _, volumeMount := range container.VolumeMounts {
								if volumeMount.Name == v.Name {
									containers = append(containers, volume.Container{Name: container.Name, MountName: volumeMount.Name, MountPath: volumeMount.MountPath})
								}
							}
						}
						// append container to pods map
						key := strings.Join([]string{pod.GetNamespace(), v.PersistentVolumeClaim.ClaimName}, ":")
						pod := volume.Pod{Namespace: pod.GetNamespace(), Name: pod.GetName(), Containers: containers}
						podsMap[key] = append(podsMap[key], pod)
					}
				}
			}
		}
	}

	// iterate through PVs once and build a map
	pvsMap := make(map[string]*v1.PersistentVolume)
	if pvs != nil {
		for _, pv := range pvs.Items {
			pvsMap[pv.GetName()] = &pv
		}
	}

	// build response object
	var resp = volume.ListResponse{}
	// loop through all the gluster volumes
	for _, gv := range glusterVolumeInfo.VolumeInfo.Volumes.Volume {
		var v = volume.Volume{}
		v.Name = gv.Name
		//gv.DistCount doesn't actually return the correct number when ReplicaCount > 1
		v.DistributionCount = gv.BrickCount / gv.ReplicaCount
		v.ReplicaCount = gv.ReplicaCount

		// get gluster volume quota
		glusterVolumeQuota, err := glusterGetter.GetQuota(gv.Name)
		if err != nil {
			return nil, err
		}
		v.Capacity = "Unknown"
		v.Available = "Unknown"
		if glusterVolumeQuota != nil && glusterVolumeQuota.VolumeQuota != nil && glusterVolumeQuota.VolumeQuota.Limit != nil {
			v.Capacity = volume.HumanFormat(glusterVolumeQuota.VolumeQuota.Limit.HardLimit)
		}
		if glusterVolumeQuota != nil && glusterVolumeQuota.VolumeQuota != nil && glusterVolumeQuota.VolumeQuota.Limit != nil {
			v.Available = volume.HumanFormat(glusterVolumeQuota.VolumeQuota.Limit.AvailSpace)
		}

		if gv.BrickCount > 0 {
			v.Bricks = make([]volume.Brick, gv.BrickCount)
			for n, gbrick := range gv.Bricks.Brick {
				brickArr := strings.Split(gbrick.Text, ":")
				v.Bricks[n] = volume.Brick{Host: brickArr[0], Path: brickArr[1]}
			}
		}

		// it is possible that all PVs were delete in kubernetes
		// set status of gluster volume to "Unknown"
		var foundPVInfo, ok = pvsMap[gv.Name]
		// this PV does not exist, maybe it was deleted?
		// set status of gluster volume to "Unknown"
		if !ok || foundPVInfo == nil {
			v.Status = "Unknown"
		} else {
			v.Labels = foundPVInfo.Labels
			v.Status = string(foundPVInfo.Status.Phase)
			if foundPVInfo.Spec.ClaimRef != nil {
				// populate claim info
				v.Claim = &volume.Claim{Namespace: foundPVInfo.Spec.ClaimRef.Namespace, Name: foundPVInfo.Spec.ClaimRef.Name}
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
