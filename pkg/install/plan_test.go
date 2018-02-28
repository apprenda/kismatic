package install

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWritePlanTemplate(t *testing.T) {
	tests := []struct {
		golden   string
		template PlanTemplateOptions
	}{
		{
			golden: "./test/plan-template.golden.yaml",
			template: PlanTemplateOptions{
				ClusterName:               "kismatic",
				InfrastructureProvisioner: "",
				EtcdNodes:                 3,
				MasterNodes:               2,
				WorkerNodes:               3,
				IngressNodes:              2,
				StorageNodes:              0,
				NFSVolumes:                0,
				AdminPassword:             "password",
			},
		},
		{
			golden: "./test/plan-template-with-storage.golden.yaml",
			template: PlanTemplateOptions{
				ClusterName:               "kubernetes",
				InfrastructureProvisioner: "aws",
				EtcdNodes:                 3,
				MasterNodes:               2,
				WorkerNodes:               3,
				IngressNodes:              2,
				StorageNodes:              2,
				NFSVolumes:                3,
				AdminPassword:             "password",
			},
		},
	}
	for _, test := range tests {
		expected, err := ioutil.ReadFile(test.golden)
		if err != nil {
			t.Fatalf("error reading golden file: %v", err)
		}
		tmp, err := ioutil.TempDir("", "ket-test-write-plan-template")
		if err != nil {
			t.Fatalf("error creating temp dir: %v", err)
		}
		file := filepath.Join(tmp, "kismatic-cluster.yaml")
		fp := &FilePlanner{file}
		if err = WritePlanTemplate(test.template, fp); err != nil {
			t.Fatalf("error writing plan template: %v", err)
		}
		wrote, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatalf("error reading plan file template: %v", err)
		}
		if !bytes.Equal(wrote, expected) {
			t.Errorf("the resulting plan file did not equal the expected plan file (%s)", test.golden)
			if c, err := exec.LookPath("diff"); err == nil {
				cmd := exec.Command(c, test.golden, file)
				result, _ := cmd.CombinedOutput()
				fmt.Println(string(result))
			}
		}
	}
}

func TestReadWithDeprecated(t *testing.T) {
	pm := &DeprecatedPackageManager{
		Enabled: true,
	}
	p := &Plan{}
	p.Features = &Features{
		PackageManager: pm,
	}
	b := false
	p.Cluster.AllowPackageInstallation = &b
	readDeprecatedFields(p)

	// features.package_manager should be set to add_ons.package_manager
	if p.AddOns.PackageManager.Disable || p.AddOns.PackageManager.Provider != "helm" {
		t.Errorf("Expected add_ons.package_manager to be read from features.package_manager")
	}
	// cluster.disable_package_installation shoule be set to cluster.allow_package_installation
	if p.Cluster.DisablePackageInstallation != true {
		t.Errorf("Expected cluster.allow_package_installation to be read from cluster.disable_package_installation")
	}
}

func TestReadWithNil(t *testing.T) {
	p := &Plan{}
	setDefaults(p)

	if p.AddOns.CNI.Provider != "calico" {
		t.Errorf("Expected add_ons.cni.provider to equal 'calico', instead got %s", p.AddOns.CNI.Provider)
	}
	if p.AddOns.CNI.Options.Calico.Mode != "overlay" {
		t.Errorf("Expected add_ons.cni.options.calico.mode to equal 'overlay', instead got %s", p.AddOns.CNI.Options.Calico.Mode)
	}

	if p.AddOns.HeapsterMonitoring.Options.Heapster.Replicas != 2 {
		t.Errorf("Expected add_ons.heapster.options.heapster.replicas to equal 2, instead got %d", p.AddOns.HeapsterMonitoring.Options.Heapster.Replicas)
	}

	if p.AddOns.HeapsterMonitoring.Options.Heapster.ServiceType != "ClusterIP" {
		t.Errorf("Expected add_ons.heapster.options.heapster.service_type to equal ClusterIP, instead got %s", p.AddOns.HeapsterMonitoring.Options.Heapster.ServiceType)
	}

	if p.AddOns.HeapsterMonitoring.Options.Heapster.Sink != "influxdb:http://heapster-influxdb.kube-system.svc:8086" {
		t.Errorf("Expected add_ons.heapster.options.heapster.service_type to equal 'influxdb:http://heapster-influxdb.kube-system.svc:8086', instead got %s", p.AddOns.HeapsterMonitoring.Options.Heapster.Sink)
	}

	if p.Cluster.Certificates.CAExpiry != defaultCAExpiry {
		t.Errorf("expected ca cert expiry to be %s, but got %s", defaultCAExpiry, p.Cluster.Certificates.CAExpiry)
	}

	if p.Cluster.Version != kubernetesVersionString {
		t.Errorf("expected cluster version to be %s, but got %s", kubernetesVersionString, p.Cluster.Version)
	}
}

func TestReadDeprecatedDashboard(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-read-deprecated-dashboard")
	if err != nil {
		t.Fatalf("error creating tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	file := filepath.Join(tmpDir, "kismatic-cluster.yaml")

	tests := []struct {
		name           string
		planStr        string
		expectDisabled bool
	}{
		{
			name:           "deprecated is set to true",
			planStr:        `{'add_ons': {'dashbard': {'disable': true}}}`,
			expectDisabled: true,
		},
		{
			name:           "deprecated is set to false",
			planStr:        `{'add_ons': {'dashbard': {'disable': false}}}`,
			expectDisabled: false,
		},
		{
			name:           "actual field is set to true",
			planStr:        `{'add_ons': {'dashboard': {'disable': true}}}`,
			expectDisabled: true,
		},
		{
			name:           "actual field is set to false",
			planStr:        `{'add_ons': {'dashboard': {'disable': false}}}`,
			expectDisabled: false,
		},
		{
			name:           "both fields are set to true",
			planStr:        `{'add_ons': {'dashboard': {'disable': true}, 'dashbard': {'disable': true}}}`,
			expectDisabled: true,
		},
		{
			name:           "both fields are set to false",
			planStr:        `{'add_ons': {'dashboard': {'disable': false}, 'dashbard': {'disable': false}}}`,
			expectDisabled: false,
		},
		{
			name:           "both are missing",
			planStr:        "",
			expectDisabled: false,
		},
		{
			name:           "deprecated is set to false, new one is set to true",
			planStr:        `{'add_ons': {'dashbard': {'disable': false}, 'dashboard': {'disable': true}}}`,
			expectDisabled: true,
		},
		{
			name:           "deprecated is set to true, new one is set to false",
			planStr:        `{'add_ons': {'dashbard': {'disable': true}, 'dashboard': {'disable': false}}}`,
			expectDisabled: false,
		},
	}

	for _, test := range tests {
		// writeFile truncates before writing, so we can reuse the file
		if err = ioutil.WriteFile(file, []byte(test.planStr), 0666); err != nil {
			t.Fatalf("error writing plan file")
		}

		planner := FilePlanner{file}
		plan, err := planner.Read()
		if err != nil {
			t.Fatalf("error reading plan file")
		}

		if plan.AddOns.Dashboard.Disable != test.expectDisabled {
			t.Errorf("name: %s: expected disabled to be %v, but got %v.", test.name, test.expectDisabled, plan.AddOns.Dashboard.Disable)
		}
	}

}

// In general, Newgroup represents the inbound Terraform NodeGroup
// And should never be labelled.
func TestMerge(t *testing.T) {
	tests := []struct {
		name          string
		OldGroup      NodeGroup
		NewGroup      NodeGroup
		ExpectedGroup NodeGroup
	}{
		{
			name: "should overwrite labels",
			OldGroup: NodeGroup{
				ExpectedCount: 4,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "3",
						IP:         "3",
						Host:       "3",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
				},
			},
			NewGroup: NodeGroup{
				ExpectedCount: 4,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label": "should NOT remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should NOT remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label": "should NOT remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should NOT remain",
							},
						},
					},
					Node{
						InternalIP: "3",
						IP:         "3",
						Host:       "3",
						Labels: map[string]string{
							"this label": "should NOT remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should NOT remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label": "should NOT remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should NOT remain",
							},
						},
					},
				},
			},
			ExpectedGroup: NodeGroup{
				ExpectedCount: 4,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "3",
						IP:         "3",
						Host:       "3",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
				},
			},
		},
		{
			name: "should merge labels",
			NewGroup: NodeGroup{
				ExpectedCount: 4,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "3",
						IP:         "3",
						Host:       "3",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
				},
			},
			OldGroup: NodeGroup{
				ExpectedCount: 4,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this other label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this other override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this other label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this other override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "3",
						IP:         "3",
						Host:       "3",
						Labels: map[string]string{
							"this other label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this other override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this other label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this other override": "should remain",
							},
						},
					},
				},
			},
			ExpectedGroup: NodeGroup{
				ExpectedCount: 4,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label":       "should remain",
							"this other label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override":       "should remain",
								"this other override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label":       "should remain",
							"this other label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override":       "should remain",
								"this other override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "3",
						IP:         "3",
						Host:       "3",
						Labels: map[string]string{
							"this label":       "should remain",
							"this other label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override":       "should remain",
								"this other override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label":       "should remain",
							"this other label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override":       "should remain",
								"this other override": "should remain",
							},
						},
					},
				},
			},
		},
		{
			name: "should scale up and maintain order",
			NewGroup: NodeGroup{
				ExpectedCount: 6,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label": "should NOT remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should NOT remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label": "should NOT remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should NOT remain",
							},
						},
					},
					Node{
						InternalIP: "3",
						IP:         "3",
						Host:       "3",
						Labels: map[string]string{
							"this label": "should NOT remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should NOT remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label": "should NOT remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should NOT remain",
							},
						},
					},
					Node{
						InternalIP: "5",
						IP:         "5",
						Host:       "5",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "6",
						IP:         "6",
						Host:       "6",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
				},
			},
			OldGroup: NodeGroup{
				ExpectedCount: 4,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "3",
						IP:         "3",
						Host:       "3",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
				},
			},
			ExpectedGroup: NodeGroup{
				ExpectedCount: 6,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "3",
						IP:         "3",
						Host:       "3",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "5",
						IP:         "5",
						Host:       "5",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "6",
						IP:         "6",
						Host:       "6",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
				},
			},
		},
		{
			name: "should scale down and maintain order",
			NewGroup: NodeGroup{
				ExpectedCount: 3,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
				},
			},
			OldGroup: NodeGroup{
				ExpectedCount: 3,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "3",
						IP:         "3",
						Host:       "3",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
				},
			},
			ExpectedGroup: NodeGroup{
				ExpectedCount: 3,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
				},
			},
		},
		{
			name: "should scale up from nothing",
			OldGroup: NodeGroup{
				ExpectedCount: 0,
				Nodes:         []Node{},
			},
			NewGroup: NodeGroup{
				ExpectedCount: 3,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
				},
			},
			ExpectedGroup: NodeGroup{
				ExpectedCount: 3,
				Nodes: []Node{
					Node{
						InternalIP: "1",
						IP:         "1",
						Host:       "1",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "2",
						IP:         "2",
						Host:       "2",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
					Node{
						InternalIP: "4",
						IP:         "4",
						Host:       "4",
						Labels: map[string]string{
							"this label": "should remain",
						},
						KubeletOptions: KubeletOptions{
							map[string]string{
								"this override": "should remain",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		merged := Merge(test.NewGroup, test.OldGroup)
		if !reflect.DeepEqual(merged, test.ExpectedGroup) {
			t.Errorf("Merge of group %v failed. Got %v, expected %v.", test.name, merged, test.ExpectedGroup)
		}
	}

}
