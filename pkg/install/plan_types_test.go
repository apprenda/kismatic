package install

import (
	"testing"
	"io/ioutil"
	yaml "gopkg.in/yaml.v2"

	"reflect"
)

func TestReadPlanFile(t *testing.T) {
	d, _ := ioutil.ReadFile("test/cluster-config.yaml")
	p := &Plan{}

	yaml.Unmarshal(d, p)

	assertEqual(t, p.Cluster.Name, "my_cluster_name")
	assertEqual(t, p.Cluster.AdminPassword, "secret_admin_password")

	assertEqual(t, p.Cluster.APIServerConfig.ConfigValues()["runtime-config"], "alpha/v1api=true,beta/v2api=true,extensions/v1beta1=true,extensions/v1beta1/networkpolicies=true");
}

func TestAPIServerRuntimeConfig(t *testing.T) {
	apiServer := APIServerConfig{
		RawData: map[string]string{
			"runtime-config": "beta/v1Option=true,beta/v2Option=false",
		},
	}

	assertEqual(t, apiServer.ConfigValues()["runtime-config"], "beta/v1Option=true,beta/v2Option=false,extensions/v1beta1=true,extensions/v1beta1/networkpolicies=true")
}

func TestAddsDefaultAPIConfigOptions(t *testing.T) {
	apiServer := APIServerConfig{
		RawData: map[string]string{},
	}

	assertEqual(t, apiServer.ConfigValues()["runtime-config"], "extensions/v1beta1=true,extensions/v1beta1/networkpolicies=true")
}

func TestCanOverrideDefaultAPIConfigOptions(t *testing.T) {
	apiServer := APIServerConfig{
		RawData: map[string]string{
			"runtime-config": "extensions/v1beta1/networkpolicies=false",
		},
	}

	assertEqual(t, apiServer.ConfigValues()["runtime-config"], "extensions/v1beta1=true,extensions/v1beta1/networkpolicies=false")
}

func TestAPIServerRuntimeConfigWithNoAPIServer(t *testing.T) {

	cluster := Cluster{}

	assertEqual(t, cluster.APIServerConfig.ConfigValues()["runtime-config"], "extensions/v1beta1=true,extensions/v1beta1/networkpolicies=true")
}

func TestAPIServerRuntimeConfigWithNoAPIServerConfigOptions(t *testing.T) {
	cluster := Cluster{
		APIServerConfig: APIServerConfig{},
	}

	assertEqual(t, cluster.APIServerConfig.ConfigValues()["runtime-config"], "extensions/v1beta1=true,extensions/v1beta1/networkpolicies=true")
}

func assertEqual(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%v != %v", a, b)
	}
}
