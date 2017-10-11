package vagrant

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
)

const vagrantCmd string = "vagrant"

func ensureVagrantOnPath() string {
	path, err := exec.LookPath(vagrantCmd)
	if err != nil {
		log.Fatal("Unable to locate vagrant on path.  It can be downloaded from http://www.vagrantup.com")
	}
	return path
}

func grabSSHConfig() string {
	cmd := exec.Command("vagrant", "ssh-config")
	bytes, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting ssh config for newly created vagrant image", err)
		os.Exit(1)
	}

	r, err := regexp.Compile(`.*IdentityFile\ (.*)`)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error compiling regex", err)
		os.Exit(1)
	}

	a := r.FindSubmatch(bytes)

	return string(a[1])
}

func vagrantUp() error {
	cmdPath := ensureVagrantOnPath()

	cmdArgs := []string{"up"}
	cmd := exec.Command(cmdPath, cmdArgs...)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	fmt.Printf("executing '%v up'\n", vagrantCmd)
	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		os.Exit(1)
	}

	cmdErr := cmd.Wait()
	if cmdErr != nil {
		return cmdErr
	}

	return nil
}
