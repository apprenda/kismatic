package cli

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/spf13/cobra"
)

type volumeAddOptions struct {
	replicaCount       int
	distributionCount  int
	storageClass       string
	allowAddress       []string
	verbose            bool
	outputFormat       string
	generatedAssetsDir string
	reclaimPolicy      string
	accessModes        string
}

// NewCmdVolumeAdd returns the command for adding storage volumes
func NewCmdVolumeAdd(out io.Writer) *cobra.Command {
	opts := volumeAddOptions{}
	cmd := &cobra.Command{
		Use:   "add CLUSTER_NAME SIZE_IN_GB [VOLUME_NAME]",
		Short: "add storage volumes to the Kubernetes cluster",
		Long: `Add storage volumes to the Kubernetes cluster.

This function requires a target cluster that has storage nodes.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return cmd.Usage()
			}
			clusterName := args[0]
			if exists, err := CheckClusterExists(clusterName); !exists {
				return err
			}
			planPath, generatedPath, _ := generateDirsFromName(clusterName)
			opts.generatedAssetsDir = generatedPath
			return doVolumeAdd(out, opts, planPath, args[1:])
		},
		Example: `  # Create a 10GB distributed and replicated volume named "storage01"
  # with StorageClass "durable". Grant access to the volume to any client with an IP
  # that starts with 10.10.
  kismatic volume add kubernetes 10 storage01 -r 2 -d 2 -c="durable" -a 10.10.*.*
		`,
	}
	cmd.Flags().IntVarP(&opts.replicaCount, "replica-count", "r", 2, "The number of times each file will be written.")
	cmd.Flags().IntVarP(&opts.distributionCount, "distribution-count", "d", 1, "This is the degree to which data will be distributed across the cluster. By default, it won't be -- each replica will receive 100% of the data. Distribution makes listing or backing up the cluster more complicated by spreading data around the cluster but makes reads and writes more performant.")
	cmd.Flags().StringVarP(&opts.storageClass, "storage-class", "c", "kismatic", "The StorageClass to present for claims in Kubernetes. Classes should identify properties of volumes in business terms, such as 'durable' or 'fast-reads'")
	cmd.Flags().StringSliceVarP(&opts.allowAddress, "allow-address", "a", nil, "Comma delimited list of address wildcards permitted access to the volume in addition to Kubernetes nodes.")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", false, "enable verbose logging")
	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "simple", `output format (options simple|raw)`)
	cmd.Flags().StringVar(&opts.reclaimPolicy, "reclaim-policy", "Retain", "Persistent volume reclaim policy (options Retain|Recycle|Delete)")
	cmd.Flags().StringVar(&opts.accessModes, "access-modes", "ReadWriteMany", "Comma-separated list of access modes for the persistent volume (options ReadWriteOnce|ReadOnlyMany|ReadWriteMany)")
	return cmd
}

func doVolumeAdd(out io.Writer, opts volumeAddOptions, planFile string, args []string) error {
	// get volume name and size from arguments
	var volumeName string
	var volumeSizeStrGB string
	switch len(args) {
	case 0:
		return errors.New("the volume size (in gigabytes) must be provided as the second argument to add")
	case 1:
		volumeSizeStrGB = args[0]
		volumeName = "kismatic-" + generateRandomString(5)
	case 2:
		volumeSizeStrGB = args[0]
		volumeName = args[1]
	default:
		return fmt.Errorf("%d arguments were provided, but add does not support more than three arguments", len(args))
	}
	volumeSizeGB, err := strconv.Atoi(volumeSizeStrGB)
	if err != nil {
		return errors.New("the volume size provided is not valid")
	}

	// setup ansible for execution
	planner := &install.FilePlanner{File: planFile}
	if !planner.PlanExists() {
		return planFileNotFoundErr{filename: planFile}
	}
	execOpts := install.ExecutorOptions{
		OutputFormat: opts.outputFormat,
		Verbose:      opts.verbose,
		// Need to refactor executor code... this will do for now as we don't need the generated assets dir in this command
		GeneratedAssetsDirectory: opts.generatedAssetsDir,
	}
	exec, err := install.NewExecutor(out, out, execOpts)
	if err != nil {
		return err
	}
	plan, err := planner.Read()
	if err != nil {
		return err
	}

	// Run validation
	vopts := &validateOpts{
		outputFormat:       opts.outputFormat,
		verbose:            opts.verbose,
		planFile:           planFile,
		skipPreFlight:      true,
		generatedAssetsDir: opts.generatedAssetsDir,
	}
	if err := doValidate(out, planner, vopts); err != nil {
		return err
	}

	v := install.StorageVolume{
		Name:              volumeName,
		SizeGB:            volumeSizeGB,
		ReplicateCount:    opts.replicaCount,
		DistributionCount: opts.distributionCount,
		StorageClass:      opts.storageClass,
		ReclaimPolicy:     opts.reclaimPolicy,
		AccessModes:       strings.Split(opts.accessModes, ","),
	}
	if opts.allowAddress != nil {
		v.AllowAddresses = opts.allowAddress
	}
	if ok, errs := install.ValidateStorageVolume(v); !ok {
		fmt.Println("The storage volume configuration is not valid:")
		for _, e := range errs {
			fmt.Printf("- %s\n", e)
		}
		return errors.New("storage volume validation failed")
	}
	if err := exec.AddVolume(plan, v); err != nil {
		return fmt.Errorf("error adding new volume: %v", err)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Successfully added the persistent volume to the kubernetes cluster.")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Use \"kubectl describe pv %s\" to view volume details.\n", v.Name)
	return nil
}

func generateRandomString(n int) string {
	// removed 1, l, o, 0 to prevent confusion
	chars := []rune("abcdefghijkmnpqrstuvwxyz23456789")
	res := make([]rune, n)
	for i := range res {
		res[i] = chars[rand.Intn(len(chars))]
	}
	return string(res)
}
