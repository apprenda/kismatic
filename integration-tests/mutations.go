package integration_tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/apprenda/kismatic/pkg/install"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mutations", func() {
	var clusterNames, clusterPaths []string
	var name, path string
	BeforeEach(func() {
		dir := setupTestWorkingDir()
		os.Chdir(dir)
		name := "test-cluster-" + generateRandomString(8)
		path := filepath.Join("clusters", name, "kismatic-cluster.yaml")
		clusterNames = append(clusterNames, name)
		clusterPaths = append(clusterPaths, path)
		fp := install.FilePlanner{File: path}
		planOpts := install.PlanTemplateOptions{
			ClusterName:               name,
			InfrastructureProvisioner: "aws",
			EtcdNodes:                 2,
			MasterNodes:               2,
			WorkerNodes:               2,
			IngressNodes:              2,
		}
		install.WritePlanTemplate(planOpts, &fp)
		skipIfAWSCredsMissing()
		cmd := exec.Command("./kismatic", "provision", name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		cmd := exec.Command("./kismatic", "destroy", name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf(`				
+++++++++++++++++++++++++++++++++++++

ERROR DESTROYING CLUSTERS ON AWS. MUST BE CLEANED UP MANUALLY.

The error: %v

+++++++++++++++++++++++++++++++++++++`, err)
		}
		Expect(err).ToNot(HaveOccurred())
	})
	Describe("Attempting to mutate a cluster", func() {
		Context("by scaling the cluster", func() {
			It("should scale up without any overrides", func() {
				name, clusterNames = clusterNames[0], clusterNames[1:]
				path, clusterPaths = clusterPaths[0], clusterPaths[1:]
				fp := &install.FilePlanner{File: path}
				plan, err := fp.Read()
				Expect(err).NotTo(HaveOccurred())
				plan.Worker.ExpectedCount++
				plan.Master.ExpectedCount++
				fp.Write(plan)
				cmd := exec.Command("./kismatic", "provision", name)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("by scaling the cluster down with the override", func() {
			It("should scale down with -allow-destruction", func() {
				name, clusterNames = clusterNames[0], clusterNames[1:]
				path, clusterPaths = clusterPaths[0], clusterPaths[1:]
				fp := &install.FilePlanner{File: path}
				plan, err := fp.Read()
				Expect(err).NotTo(HaveOccurred())
				plan.Worker.ExpectedCount--
				fp.Write(plan)
				cmd := exec.Command("./kismatic", "provision", name, "--allow-destruction")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("by scaling the cluster down without the override", func() {
			It("should fail to scale down", func() {
				name, clusterNames = clusterNames[0], clusterNames[1:]
				path, clusterPaths = clusterPaths[0], clusterPaths[1:]
				fp := &install.FilePlanner{File: path}
				plan, err := fp.Read()
				Expect(err).NotTo(HaveOccurred())
				plan.Worker.ExpectedCount--
				fp.Write(plan)
				cmd := exec.Command("./kismatic", "provision", name)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
