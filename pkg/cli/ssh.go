package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"

	"github.com/apprenda/kismatic/pkg/install"
	"github.com/apprenda/kismatic/pkg/util"
	"github.com/spf13/cobra"
)

type sshOpts struct {
	planFilename string
	host         string
	pty          bool
	arguments    []string
}

// NewCmdSSH returns an ssh shell
func NewCmdSSH(out io.Writer) *cobra.Command {
	opts := &sshOpts{}

	cmd := &cobra.Command{
		Use:   "ssh CLUSTER_NAME HOST [commands]",
		Short: "ssh into a node in the cluster",
		Long: `ssh into a node in the cluster

HOST must be one of the following:
- A hostname or IP defined in the plan file
- An alias: master, etcd, worker, ingress or storage. This will ssh into the first defined node of that type.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return cmd.Usage()
			}
			clusterName := args[0]
			if exists, err := CheckClusterExists(clusterName); !exists {
				return err
			}
			planPath, _, _ := generateDirsFromName(clusterName)
			// get optional arguments
			if len(args) > 2 {
				opts.arguments = args[1:]
			}
			opts.planFilename = planPath
			opts.host = args[1]

			planner := &install.FilePlanner{File: opts.planFilename}
			// Check if plan file exists
			if !planner.PlanExists() {
				return planFileNotFoundErr{filename: opts.planFilename}
			}

			err := doSSH(out, planner, opts)
			// 130 = terminated by Control-C, so not an actual error
			if err != nil && !strings.Contains(err.Error(), "130") {
				return fmt.Errorf("SSH error %q: %v", opts.host, err)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.pty, "pty", "t", false, "force PTY \"-t\" flag on the SSH connection")

	return cmd
}

func doSSH(out io.Writer, planner install.Planner, opts *sshOpts) error {
	plan, err := planner.Read()
	if err != nil {
		return fmt.Errorf("error reading plan file: %v", err)
	}

	// find node
	con, err := plan.GetSSHConnection(opts.host)
	if err != nil {
		return err
	}

	// validate SSH access to node
	ok, errs := install.ValidateSSHConnection(con, "")
	if !ok {
		util.PrintValidationErrors(out, errs)
		return fmt.Errorf("cannot validate SSH connection to node %q", opts.host)
	}

	client, err := ssh.NewClient(con.Node.IP, con.SSHConfig.Port, con.SSHConfig.User, con.SSHConfig.Key)
	if err != nil {
		return fmt.Errorf("error creating SSH client: %v", err)
	}

	if err = client.Shell(opts.pty, opts.arguments...); err != nil {
		return fmt.Errorf("error running command: %v", err)
	}

	return err
}
