package cli

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

const seedRegistryLong = `
Seed a registry with the container images required by KET during the installation
or upgrade of your Kubernetes cluster.

The docker command line client must be installed on this machine to be able 
to push the images to the registry.

The location of the registry is obtained from the plan file by default. If you
don't have a plan file, you can pass the location of the registry using the 
--server flag. The server specified through the flag takes precedence over the 
one defined in the plan file.

If you want to further control how your registry is seeded, or if you are only
interested in the list of all images that can be used in a KET installation, you
may use the --list-only flag.
`

const imageManifestFile = "./ansible/playbooks/group_vars/container_images.yaml"

type seedRegistryOptions struct {
	listOnly       bool
	verbose        bool
	planFile       string
	registryServer string
}

type imageManifest struct {
	OfficialImages map[string]image `yaml:"official_images"`
}

type image struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func (i image) String() string {
	return fmt.Sprintf("%s:%s", i.Name, i.Version)
}

// NewCmdSeedRegistry returns the command for seeding a container image registry
// with the images required by KET
func NewCmdSeedRegistry(stdout, stderr io.Writer) *cobra.Command {
	var options seedRegistryOptions
	cmd := &cobra.Command{
		Use:   "seed-registry",
		Short: "seed a registry with the container images required by KET",
		Long:  seedRegistryLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return cmd.Usage()
			}
			if options.listOnly {
				return doListImages(stdout, options, imageManifestFile)
			}
			return doSeedRegistry(stdout, stderr, options, imageManifestFile)
		},
	}
	cmd.Flags().BoolVar(&options.listOnly, "list-only", false, "when true, the images will only be listed but not pushed to the registry")
	cmd.Flags().BoolVar(&options.verbose, "verbose", false, "enable verbose logging")
	cmd.Flags().StringVar(&options.registryServer, "server", "", "set to the location of the registry server, without the protocol (e.g. localhost:5000)")
	addPlanFileFlag(cmd.Flags(), &options.planFile)
	return cmd
}

func doListImages(out io.Writer, options seedRegistryOptions, imageManifestFile string) error {
	im, err := readImageManifest()
	if err != nil {
		return err
	}
	for _, img := range im.OfficialImages {
		fmt.Fprintf(out, "%s\n", img)
	}
	return nil
}

func doSeedRegistry(stdout, stderr io.Writer, options seedRegistryOptions, imageManifestFile string) error {
	util.PrintHeader(stdout, "Seed Container Image Registry", '=')

	// Validate that docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return errors.New("Did not find docker installed on this node. The docker CLI must be available for seeding the registry.")
	}

	// Figure out the registry we are to seed
	// The registry specified through the command-line flag takes precedence
	// over the one defined in the plan file.
	server := options.registryServer
	if server == "" {
		// we need to get the server from the plan file
		planner := install.FilePlanner{File: options.planFile}
		if !planner.PlanExists() {
			util.PrettyPrintErr(stdout, "Reading installation plan file [ERROR]")
			fmt.Fprintln(stdout, `Run "kismatic install plan" to generate it or use the "--server" option`)
			return fmt.Errorf("plan does not exist")
		}
		plan, err := planner.Read()
		if err != nil {
			util.PrettyPrintErr(stdout, "Reading installation plan file %q", options.planFile)
			return fmt.Errorf("error reading plan file: %v", err)
		}
		util.PrettyPrintOk(stdout, "Reading installation plan file %q", options.planFile)
		// Validate the registry info in the plan file
		errs := []error{}
		if plan.DockerRegistry.Address == "" {
			errs = append(errs, errors.New("The private registry's address must be set in the plan file."))
		}
		if plan.DockerRegistry.Port == 0 {
			errs = append(errs, errors.New("The private registry's port must be set in the plan file."))
		}
		if plan.DockerRegistry.Port < 1 || plan.DockerRegistry.Port > 65535 {
			errs = append(errs, fmt.Errorf("The private registry port '%d' provided in the plan file is not valid.", plan.DockerRegistry.Port))
		}
		if len(errs) > 0 {
			util.PrettyPrintErr(stdout, "Validating registry configured in plan file")
			util.PrintValidationErrors(stdout, errs)
			return errors.New("Invalid registry configuration found in plan file")
		}
		server = fmt.Sprintf("%s:%d", plan.DockerRegistry.Address, plan.DockerRegistry.Port)
	}

	im, err := readImageManifest()
	if err != nil {
		return err
	}

	// Seed the registry with the images
	n := len(im.OfficialImages)
	i := 1
	for _, img := range im.OfficialImages {
		l := fmt.Sprintf("(%d/%d) Seeding %s ", i, n, img)
		pad := 80 - len(l)
		if pad < 0 {
			pad = 0
		}
		fmt.Fprintf(stdout, l+strings.Repeat(" ", pad))
		if err := seedImage(stdout, stderr, img, server, options.verbose); err != nil {
			return fmt.Errorf("Error seeding image %q: %v", img, err)
		}
		util.PrintOkln(stdout)
		i++
	}

	util.PrintColor(stdout, util.Green, "\nThe registry %q was seeded successfully.\n", server)
	fmt.Fprintln(stdout)
	return nil
}

func seedImage(stdout, stderr io.Writer, img image, registry string, verbose bool) error {
	runDockerCmd := func(args ...string) error {
		command := exec.Command("docker", args...)
		command.Stderr = stderr
		if verbose {
			command.Stdout = stdout
		}
		return command.Run()
	}

	// pull
	if err := runDockerCmd("pull", img.String()); err != nil {
		return err
	}
	// tag
	privateImgTag := fmt.Sprintf("%s/%s", registry, img)
	if err := runDockerCmd("tag", img.String(), privateImgTag); err != nil {
		return err
	}
	// push
	if err := runDockerCmd("push", privateImgTag); err != nil {
		return err
	}
	return nil
}

func readImageManifest() (imageManifest, error) {
	im := imageManifest{}
	imBytes, err := ioutil.ReadFile(imageManifestFile)
	if err != nil {
		return im, fmt.Errorf("Error reading the list of images: %v", err)
	}
	if err := yaml.Unmarshal(imBytes, &im); err != nil {
		return im, fmt.Errorf("Error unmarshalling the list of images: %v", err)
	}
	return im, nil
}
