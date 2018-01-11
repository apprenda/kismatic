package integration_tests

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/apprenda/kismatic/pkg/install"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mutations", func() {
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)

		fp := install.FilePlanner{File: "kismatic-cluster.yaml"}
		planOpts := install.PlanTemplateOptions{
			ClusterName:               "test-cluster-" + generateRandomString(8),
			InfrastructureProvisioner: "aws",
			EtcdNodes:                 2,
			MasterNodes:               2,
			WorkerNodes:               2,
			IngressNodes:              2,
		}
		install.WritePlanTemplate(planOpts, &fp)
		skipIfAWSCredsMissing()
		cmd := exec.Command("./kismatic", "install", "provision")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		cmd := exec.Command("./kismatic", "install", "destroy")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf(`+++++++++++++++++++++++++++++++++++++

ERROR DESTROYING CLUSTERS ON AWS. MUST BE CLEANED UP MANUALLY.

The error: %v

+++++++++++++++++++++++++++++++++++++`, err)
		}
		Expect(err).ToNot(HaveOccurred())
	})
	Describe("Attempting to mutate a cluster", func() {
		Context("by scaling the cluster up", func() {
			It("should scale up without any overrides", func() {
				planFileName := "kismatic-cluster.yaml"
				fp := &install.FilePlanner{File: planFileName}
				plan, err := fp.Read()
				Expect(err).NotTo(HaveOccurred())
				plan.Worker.ExpectedCount++
				plan.Master.ExpectedCount++
				fp.Write(plan)
				cmd := exec.Command("./kismatic", "install", "provision")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("by scaling the cluster down with the override", func() {
			It("should scale down with -allow-destruction", func() {
				planFileName := "kismatic-cluster.yaml"
				fp := &install.FilePlanner{File: planFileName}
				plan, err := fp.Read()
				Expect(err).NotTo(HaveOccurred())
				plan.Worker.ExpectedCount--
				fp.Write(plan)
				cmd := exec.Command("./kismatic", "install", "provision", "--allow-destruction")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("by scaling the cluster down without the override", func() {
			It("should fail to scale down", func() {
				planFileName := "kismatic-cluster.yaml"
				fp := &install.FilePlanner{File: planFileName}
				plan, err := fp.Read()
				Expect(err).NotTo(HaveOccurred())
				plan.Worker.ExpectedCount--
				fp.Write(plan)
				cmd := exec.Command("./kismatic", "install", "provision")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
