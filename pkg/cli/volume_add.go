package cli

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"

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
}

// NewCmdVolumeAdd returns the command for adding storage volumes
func NewCmdVolumeAdd(out io.Writer, planFile *string) *cobra.Command {
	opts := volumeAddOptions{}
	cmd := &cobra.Command{
		Use:   "add size_in_gigabytes [volume name]",
		Short: "add storage volumes to the Kubernetes cluster",
		Long: `Add storage volumes to the Kubernetes cluster.

This function requires a target cluster that has storage nodes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doVolumeAdd(out, opts, *planFile, args)
		},
		Example: `  Create a distributed, replicated volume,
  named "storage01", a kubernetes StorageClass of "durable", with a 10 GB quota,
  grant access to any IP address starting with 10.10.
  kismatic volume add 10 storage01 -r 2 -d 2 -c="durable" -a 10.10.*.*
		`,
	}
	cmd.Flags().IntVarP(&opts.replicaCount, "replica-count", "r", 2, "The number of times each file will be written.")
	cmd.Flags().IntVarP(&opts.distributionCount, "distribution-count", "d", 1, "This is the degree to which data will be distributed across the cluster. By default, it won't be -- each replica will receive 100% of the data. Distribution makes listing or backing up the cluster more complicated by spreading data around the cluster but makes reads and writes more performant.")
	cmd.Flags().StringVarP(&opts.storageClass, "storage-class", "c", "kismatic", "The StorageClass to present for claims in Kubernetes. Classes should identify properties of volumes in business terms, such as 'durable' or 'fast-reads'")
	cmd.Flags().StringSliceVarP(&opts.allowAddress, "allow-address", "a", nil, "Comma delimited list of address wildcards permitted access to the volume in addition to Kubernetes nodes.")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", false, "enable verbose logging")
	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "simple", `output format (options "simple"|"raw")`)
	cmd.Flags().StringVar(&opts.generatedAssetsDir, "generated-assets-dir", "generated", "path to the directory where assets generated during the installation process will be stored")
	return cmd
}

func doVolumeAdd(out io.Writer, opts volumeAddOptions, planFile string, args []string) error {
	// Setup ansible
	planner := &install.FilePlanner{File: planFile}
	if !planner.PlanExists() {
		return fmt.Errorf("plan file not found at %q", planFile)
	}

	// verify command
	var volumeName string
	var volumeSizeStrGB string
	switch len(args) {
	case 0:
		return errors.New("the volume size (in gigabytes) must be provided as the first argument to add")
	case 1:
		volumeSizeStrGB = args[0]
		volumeName = "kismatic-" + generateRandomString(5)
	case 2:
		volumeSizeStrGB = args[0]
		volumeName = args[1]
	default:
		return fmt.Errorf("%d arguments were provided, but add does not support more than two arguments", len(args))
	}
	volumeSizeGB, err := strconv.Atoi(volumeSizeStrGB)
	if err != nil {
		return errors.New("the volume size provided is not valid")
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
