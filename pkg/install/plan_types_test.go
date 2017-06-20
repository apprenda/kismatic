package install

import (
	"testing"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func TestCanReadAPIServerOverrides(t *testing.T) {
	d, _ := ioutil.ReadFile("test/cluster-config.yaml")
	p := &Plan{}
	yaml.Unmarshal(d, p)

	assertEqual(t, p.Cluster.APIServerOptions.Overrides["runtime-config"], "beta/v2api=true,alpha/v1api=true")
}