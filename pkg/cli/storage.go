package cli

import (
	"github.com/spf13/cobra"
	"io"
	"fmt"
	"math/rand"
	"errors"
	"strconv"
	"github.com/apprenda/kismatic/pkg/install"
)

// NewCmdStorage returns the storage command
func NewCmdStorage(out io.Writer) *cobra.Command {
	var planFile string
	cmd := &cobra.Command{
		Use: "storage",
		Short: "manage storage on your Kubernetes cluster",
		RunE:func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	addPlanFileFlag(cmd.PersistentFlags(), &planFile)
	cmd.AddCommand(NewCmdStorageAddVolume(out, &planFile))
	return cmd
}

type addVolumeOptions struct {
	replicaCount int
	distributionCount int
	allowAddress []string
	verbose bool
	outputFormat string
}

func NewCmdStorageAddVolume(out io.Writer, planFile *string) *cobra.Command {
	opts := addVolumeOptions{}
	cmd := &cobra.Command{
		Use: "add-volume size_in_gigabytes [volume name]",
		Short: "add storage volumes to the Kubernetes cluster",
		Long: `Add storage volumes to the Kubernetes cluster.

This function requires a target cluster that has storage nodes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doStorageAddVolume(out, opts, *planFile, args)
		},
		Example: `# Create a distributed, replicated volume named kismatic with a 10 GB quota and
		# grant access to any IP address starting with 10.10.
		kismatic storage new-volume -r 2 -d 2 -a 10.10.*.* 10 kismatic
		`,
	}
	cmd.Flags().IntVarP(&opts.replicaCount, "replica-count", "r", 2, "The number of times each file will be written.")
	cmd.Flags().IntVarP(&opts.distributionCount, "distribution-count", "d", 1, "This is the degree to which data will be distributed across the cluster. By default, it won't be -- each replica will receive 100% of the data. Distribution makes listing or backing up the cluster more complicated by spreading data around the cluster but makes reads and writes more performant.")
	cmd.Flags().StringSliceVarP(&opts.allowAddress, "allow-address", "a", nil, "Comma delimited list of address wildcards permitted access to the volume in addition to Kubernetes nodes.")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", false, "enable verbose logging")
	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "simple", `output format (options "simple"|"raw")`)
	return cmd
}

func doStorageAddVolume(out io.Writer, opts addVolumeOptions, planFile string, args []string) error {
	var volumeName string
	var volumeSizeStrGB string
	switch len(args) {
	case 0:
		return errors.New("the volume size (in gigabytes) must be provided as the first argument to add-volume")
	case 1:
		volumeSizeStrGB = args[0]
		volumeName = "kismatic-" + generateRandomString(5)
	case 2:
		volumeSizeStrGB = args[0]
		volumeName = args[1]
	default:
		return fmt.Errorf("%d arguments were provided, but add-volume does not support more than two arguments", len(args))
	}
	volumeSizeGB, err := strconv.Atoi(volumeSizeStrGB)
	if err != nil {
		return errors.New("the volume size provided is not valid")
	}
	// Setup ansible
	planner := &install.FilePlanner{File: planFile}
	if !planner.PlanExists() {
		return fmt.Errorf("Plan file not found at %q", planFile)
	}
	execOpts := install.ExecutorOptions{
		OutputFormat: opts.outputFormat,
		Verbose: opts.verbose,
		// Need to refactor executor code... this will do for now as we don't need the generated assets dir in this command
		GeneratedAssetsDirectory: "generated",
	}
	exec, err := install.NewExecutor(out, out, execOpts)
	if err != nil {
		return err
	}
	plan, err := planner.Read()
	if err != nil {
		return err
	}
	v := install.StorageVolume{
		Name: volumeName,
		SizeGB: volumeSizeGB,
		ReplicateCount: opts.replicaCount,
		DistributionCount: opts.distributionCount,
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
	exec.AddVolume(plan, v)
	return nil
}

func generateRandomString(n int) string {
	// removed 1, l, o, 0 and l to prevent confusion
	chars := []rune("abcdefghijkmnpqrstuvwxyz23456789")
	res := make([]rune, n)
	for i := range res {
		res[i] = chars[rand.Intn(len(chars))]
	}
	return string(res)
}