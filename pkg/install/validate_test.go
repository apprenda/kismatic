package install

import (
	"fmt"
	"testing"
)

var validPlan = Plan{
	Cluster: Cluster{
		Name:          "test",
		AdminPassword: "password",
		Networking: NetworkConfig{
			Type:             "overlay",
			PodCIDRBlock:     "172.16.0.0/16",
			ServiceCIDRBlock: "172.17.0.0/16",
		},
		Certificates: CertsConfig{
			Expiry: "17250h",
		},
		SSH: SSHConfig{
			User: "root",
			Key:  "/bin/sh",
			Port: 22,
		},
	},
	Etcd: NodeGroup{
		ExpectedCount: 1,
		Nodes: []Node{
			{
				Host: "etcd01",
				IP:   "192.168.205.10",
			},
		},
	},
	Master: MasterNodeGroup{
		ExpectedCount: 1,
		Nodes: []Node{
			{
				Host: "etcd01",
				IP:   "192.168.205.10",
			},
		},
		LoadBalancedFQDN:      "test",
		LoadBalancedShortName: "test",
	},
	Worker: NodeGroup{
		ExpectedCount: 1,
		Nodes: []Node{
			{
				Host: "etcd01",
				IP:   "192.168.205.10",
			},
		},
	},
	Ingress: OptionalNodeGroup{
		ExpectedCount: 1,
		Nodes: []Node{
			{
				Host: "etcd01",
				IP:   "192.168.205.10",
			},
		},
	},
}

func assertInvalidPlan(t *testing.T, p Plan) {
	valid, _ := ValidatePlan(&p)
	if valid {
		t.Errorf("expected invalid, but got valid")
	}
}

func TestValidateBlankPlan(t *testing.T) {
	p := Plan{}
	assertInvalidPlan(t, p)
}

func TestValidateValidPlan(t *testing.T) {
	p := validPlan
	valid, errs := ValidatePlan(&p)
	if !valid {
		t.Errorf("expected valid, but got invalid")
	}
	fmt.Println(errs)
}

func TestValidatePlanInvalidNetworkOption(t *testing.T) {
	p := validPlan
	p.Cluster.Networking.Type = "foo"
	assertInvalidPlan(t, p)
}

func TestValidatePlanEmptyPodCIDR(t *testing.T) {
	p := validPlan
	p.Cluster.Networking.PodCIDRBlock = ""
	assertInvalidPlan(t, p)
}

func TestValidatePlanInvalidPodCIDR(t *testing.T) {
	p := validPlan
	p.Cluster.Networking.PodCIDRBlock = "foo"
	assertInvalidPlan(t, p)
}

func TestValidatePlanEmptyServicesCIDR(t *testing.T) {
	p := validPlan
	p.Cluster.Networking.ServiceCIDRBlock = ""
	assertInvalidPlan(t, p)
}

func TestValidatePlanInvalidServicesCIDR(t *testing.T) {
	p := validPlan
	p.Cluster.Networking.ServiceCIDRBlock = "foo"
	assertInvalidPlan(t, p)
}

func TestValidatePlanEmptyPassword(t *testing.T) {
	p := validPlan
	p.Cluster.AdminPassword = ""
	assertInvalidPlan(t, p)
}

func TestValidatePlanEmptyCertificatesExpiry(t *testing.T) {
	p := validPlan
	p.Cluster.Certificates.Expiry = ""
	assertInvalidPlan(t, p)
}

func TestValidatePlanInvalidCertExpiry(t *testing.T) {
	p := validPlan
	p.Cluster.Certificates.Expiry = "foo"
	assertInvalidPlan(t, p)
}

func TestValidatePlanEmptySSHUser(t *testing.T) {
	p := validPlan
	p.Cluster.SSH.User = ""
	assertInvalidPlan(t, p)
}

func TestValidatePlanEmptySSHKey(t *testing.T) {
	p := validPlan
	p.Cluster.SSH.Key = ""
	assertInvalidPlan(t, p)
}

func TestValidatePlanNonExistentSSHKey(t *testing.T) {
	p := validPlan
	p.Cluster.SSH.Key = "/foo"
	assertInvalidPlan(t, p)
}

func TestValidatePlanNegativeSSHPort(t *testing.T) {
	p := validPlan
	p.Cluster.SSH.Port = -1
	assertInvalidPlan(t, p)
}

func TestValidatePlanEmptyLoadBalancedFQDN(t *testing.T) {
	p := validPlan
	p.Master.LoadBalancedFQDN = ""
	assertInvalidPlan(t, p)
}

func TestValidatePlanEmptyLoadBalancedShortName(t *testing.T) {
	p := validPlan
	p.Master.LoadBalancedShortName = ""
	assertInvalidPlan(t, p)
}

func TestValidatePlanNoEtcdNodes(t *testing.T) {
	p := validPlan
	p.Etcd.ExpectedCount = 0
	p.Etcd.Nodes = []Node{}
	assertInvalidPlan(t, p)
}

func TestValidatePlanNoMasterNodes(t *testing.T) {
	p := validPlan
	p.Master.ExpectedCount = 0
	p.Master.Nodes = []Node{}
	assertInvalidPlan(t, p)
}

func TestValidatePlanNoWorkerNodes(t *testing.T) {
	p := validPlan
	p.Worker.ExpectedCount = 0
	p.Worker.Nodes = []Node{}
	assertInvalidPlan(t, p)
}

func TestValidatePlanEtcdNodesMismatch(t *testing.T) {
	p := validPlan
	p.Etcd.ExpectedCount = 100
	assertInvalidPlan(t, p)
}

func TestValidatePlanMasterNodesMismatch(t *testing.T) {
	p := validPlan
	p.Master.ExpectedCount = 100
	assertInvalidPlan(t, p)
}

func TestValidatePlanWorkerNodesMismatch(t *testing.T) {
	p := validPlan
	p.Worker.ExpectedCount = 100
	assertInvalidPlan(t, p)
}

func TestValidatePlanUnexpectedEtcdNodes(t *testing.T) {
	p := validPlan
	p.Etcd.ExpectedCount = 1
	p.Etcd.Nodes = []Node{
		{
			Host: "etcd01",
			IP:   "192.168.205.10",
		},
		{
			Host: "etcd02",
			IP:   "192.168.205.11",
		},
	}
	assertInvalidPlan(t, p)
}

func TestValidatePlanUnexpectedMasterNodes(t *testing.T) {
	p := validPlan
	p.Master.ExpectedCount = 1
	p.Master.Nodes = []Node{
		{
			Host: "master01",
			IP:   "192.168.205.10",
		},
		{
			Host: "master02",
			IP:   "192.168.205.11",
		},
	}
	assertInvalidPlan(t, p)
}

func TestValidatePlanUnexpectedWorkerNodes(t *testing.T) {
	p := validPlan
	p.Worker.ExpectedCount = 1
	p.Worker.Nodes = []Node{
		{
			Host: "worker01",
			IP:   "192.168.205.10",
		},
		{
			Host: "worker02",
			IP:   "192.168.205.11",
		},
	}
	assertInvalidPlan(t, p)
}

func TestValidatePlanNoIngress(t *testing.T) {
	p := validPlan
	p.Ingress.ExpectedCount = 0
	p.Ingress.Nodes = []Node{}
	valid, _ := ValidatePlan(&p)
	if !valid {
		t.Errorf("expected valid, but got invalid")
	}
}

func TestValidatePlanIngressExpected(t *testing.T) {
	p := validPlan
	p.Ingress.ExpectedCount = 1
	p.Ingress.Nodes = []Node{}
	assertInvalidPlan(t, p)
}

func TestValidatePlanIngressProvidedNotExpected(t *testing.T) {
	p := validPlan
	p.Ingress.ExpectedCount = 0
	p.Ingress.Nodes = []Node{
		{
			Host: "ingress",
			IP:   "192.168.205.10",
		},
	}
	assertInvalidPlan(t, p)
}
