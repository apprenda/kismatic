package cli

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"testing"

	"github.com/apprenda/kismatic/pkg/data"

	"k8s.io/kubernetes/pkg/api/v1"
)

type fakePodGetter struct {
	podList     []byte
	isNil       bool
	shouldError bool
}

type fakePVGetter struct {
	pvList      []byte
	isNil       bool
	shouldError bool
}

type fakeGlusterGetter struct {
	glusterVolumeList []byte
	glusterQuotas     map[string][]byte
	isNil             bool
	shouldError       bool
}

func (g fakePodGetter) Get() (*v1.PodList, error) {
	if g.isNil {
		return nil, nil
	}
	if g.shouldError {
		return nil, fmt.Errorf("error")
	}
	var pods v1.PodList
	if strings.Contains(string(g.podList), "No resources found") {
		return nil, nil
	}
	err := json.Unmarshal(g.podList, &pods)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling pod data: %v", err)
	}

	return &pods, nil
}

func (g fakePVGetter) Get() (*v1.PersistentVolumeList, error) {
	if g.isNil {
		return nil, nil
	}
	if g.shouldError {
		return nil, fmt.Errorf("error")
	}
	if strings.Contains(string(g.pvList), "No resources found") {
		return nil, nil
	}
	var pvs v1.PersistentVolumeList
	err := json.Unmarshal(g.pvList, &pvs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling persistent volume data: %v", err)
	}

	return &pvs, nil
}

func (g fakeGlusterGetter) GetVolumes() (*data.GlusterVolumeInfoCliOutput, error) {
	if g.isNil {
		return nil, nil
	}
	if g.shouldError {
		return nil, fmt.Errorf("error")
	}
	var glusterVolumeInfo data.GlusterVolumeInfoCliOutput
	err := xml.Unmarshal(g.glusterVolumeList, &glusterVolumeInfo)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling volume info data: %v", err)
	}

	return &glusterVolumeInfo, nil
}
func (g fakeGlusterGetter) GetQuota(volume string) (*data.GlusterVolumeQuotaCliOutput, error) {
	if g.isNil {
		return nil, nil
	}
	if g.shouldError {
		return nil, fmt.Errorf("error")
	}
	var glusterVolumeQuota data.GlusterVolumeQuotaCliOutput
	err := xml.Unmarshal(g.glusterQuotas[volume], &glusterVolumeQuota)
	if err != nil {
		return nil, nil
	}
	return &glusterVolumeQuota, nil
}

type volumeListTester struct {
	podGetter          fakePodGetter
	pvGetter           fakePVGetter
	glusterGetter      fakeGlusterGetter
	glusterQuotas      map[string][]byte
	volumesCount       int
	boundVolumeCount   int
	claimedVolumeCount int
	shouldBeNil        bool
	shouldError        bool
}

func TestBuildResponse(t *testing.T) {
	tests := []volumeListTester{
		volumeListTester{
			glusterGetter: fakeGlusterGetter{isNil: true},
			shouldBeNil:   true,
		},
		volumeListTester{
			glusterGetter: fakeGlusterGetter{shouldError: true},
			shouldBeNil:   true,
			shouldError:   true,
		},
		volumeListTester{
			podGetter:   fakePodGetter{shouldError: true},
			shouldBeNil: true,
			shouldError: true,
		},
		volumeListTester{
			pvGetter:    fakePVGetter{shouldError: true},
			shouldBeNil: true,
			shouldError: true,
		},
		volumeListTester{
			glusterGetter: fakeGlusterGetter{glusterVolumeList: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
		<cliOutput>
		  <opRet>0</opRet>
		  <opErrno>0</opErrno>
		  <opErrstr/>
		  <volInfo>
		    <volumes>
		      <count>0</count>
		    </volumes>
		  </volInfo>
		</cliOutput>`)},
			pvGetter: fakePVGetter{pvList: []byte(`No resources found.
		{
		    "apiVersion": "v1",
		    "items": [],
		    "kind": "List",
		    "metadata": {},
		    "resourceVersion": "",
		    "selfLink": ""
		}`)},
			podGetter: fakePodGetter{podList: []byte(`No resources found.
		{
		    "apiVersion": "v1",
		    "items": [],
		    "kind": "List",
		    "metadata": {},
		    "resourceVersion": "",
		    "selfLink": ""
		}`)},
			shouldBeNil: true,
			shouldError: false,
		},
		volumeListTester{
			glusterGetter: fakeGlusterGetter{glusterVolumeList: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cliOutput>
  <opRet>0</opRet>
  <opErrno>0</opErrno>
  <opErrstr/>
  <volInfo>
    <volumes>
      <volume>
        <name>storage1</name>
        <id>f7803d45-f974-4832-81d1-1aa9db4be522</id>
        <status>1</status>
        <statusStr>Started</statusStr>
        <snapshotCount>0</snapshotCount>
        <brickCount>1</brickCount>
        <distCount>1</distCount>
        <stripeCount>1</stripeCount>
        <replicaCount>1</replicaCount>
        <arbiterCount>0</arbiterCount>
        <disperseCount>0</disperseCount>
        <redundancyCount>0</redundancyCount>
        <type>0</type>
        <typeStr>Distribute</typeStr>
        <transport>0</transport>
        <xlators/>
        <bricks>
          <brick uuid="3cf478d7-27da-4382-8e9f-44cc72a7beb2">ip-10-0-3-199:/data/storage1<name>ip-10-0-3-199:/data/storage1</name><hostUuid>3cf478d7-27da-4382-8e9f-44cc72a7beb2</hostUuid><isArbiter>0</isArbiter></brick>
        </bricks>
        <optCount>7</optCount>
        <options>
          <option>
            <name>features.quota-deem-statfs</name>
            <value>on</value>
          </option>
          <option>
            <name>features.inode-quota</name>
            <value>on</value>
          </option>
          <option>
            <name>features.quota</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.rpc-auth-allow</name>
            <value>172.16.0.0/16,10.0.3.194,10.0.3.230,10.0.3.204,10.0.3.65,10.0.3.118,10.0.3.75,10.0.3.199</value>
          </option>
          <option>
            <name>transport.address-family</name>
            <value>inet</value>
          </option>
          <option>
            <name>performance.readdir-ahead</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.disable</name>
            <value>off</value>
          </option>
        </options>
      </volume>
      <count>1</count>
    </volumes>
  </volInfo>
</cliOutput>`),
				glusterQuotas: map[string][]byte{"storage1": []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cliOutput>
  <opRet>0</opRet>
  <opErrno>0</opErrno>
  <opErrstr/>
  <volQuota>
    <limit>
      <path>/</path>
      <hard_limit>1073741824</hard_limit>
      <soft_limit_percent>80%</soft_limit_percent>
      <soft_limit_value>858993459</soft_limit_value>
      <used_space>0</used_space>
      <avail_space>1073741824</avail_space>
      <sl_exceeded>No</sl_exceeded>
      <hl_exceeded>No</hl_exceeded>
    </limit>
  </volQuota>
</cliOutput>`)},
			},
			pvGetter: fakePVGetter{pvList: []byte(`No resources found.
{
    "apiVersion": "v1",
    "items": [],
    "kind": "List",
    "metadata": {},
    "resourceVersion": "",
    "selfLink": ""
}`)},
			podGetter: fakePodGetter{podList: []byte(`No resources found.
{
    "apiVersion": "v1",
    "items": [],
    "kind": "List",
    "metadata": {},
    "resourceVersion": "",
    "selfLink": ""
}`)},
			shouldBeNil:  false,
			shouldError:  false,
			volumesCount: 1,
		},
		volumeListTester{
			glusterGetter: fakeGlusterGetter{glusterVolumeList: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cliOutput>
  <opRet>0</opRet>
  <opErrno>0</opErrno>
  <opErrstr/>
  <volInfo>
    <volumes>
      <volume>
        <name>storage1</name>
        <id>f7803d45-f974-4832-81d1-1aa9db4be522</id>
        <status>1</status>
        <statusStr>Started</statusStr>
        <snapshotCount>0</snapshotCount>
        <brickCount>1</brickCount>
        <distCount>1</distCount>
        <stripeCount>1</stripeCount>
        <replicaCount>1</replicaCount>
        <arbiterCount>0</arbiterCount>
        <disperseCount>0</disperseCount>
        <redundancyCount>0</redundancyCount>
        <type>0</type>
        <typeStr>Distribute</typeStr>
        <transport>0</transport>
        <xlators/>
        <bricks>
          <brick uuid="3cf478d7-27da-4382-8e9f-44cc72a7beb2">ip-10-0-3-199:/data/storage1<name>ip-10-0-3-199:/data/storage1</name><hostUuid>3cf478d7-27da-4382-8e9f-44cc72a7beb2</hostUuid><isArbiter>0</isArbiter></brick>
        </bricks>
        <optCount>7</optCount>
        <options>
          <option>
            <name>features.quota-deem-statfs</name>
            <value>on</value>
          </option>
          <option>
            <name>features.inode-quota</name>
            <value>on</value>
          </option>
          <option>
            <name>features.quota</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.rpc-auth-allow</name>
            <value>172.16.0.0/16,10.0.3.194,10.0.3.230,10.0.3.204,10.0.3.65,10.0.3.118,10.0.3.75,10.0.3.199</value>
          </option>
          <option>
            <name>transport.address-family</name>
            <value>inet</value>
          </option>
          <option>
            <name>performance.readdir-ahead</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.disable</name>
            <value>off</value>
          </option>
        </options>
      </volume>
      <count>1</count>
    </volumes>
  </volInfo>
</cliOutput>`),
				glusterQuotas: map[string][]byte{"storage1": []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cliOutput>
  <opRet>-1</opRet>
  <opErrno>30800</opErrno>
  <opErrstr>Volume storage01 does not exist</opErrstr>
  <cliOp>volQuota</cliOp>
</cliOutput>`)},
			},
			pvGetter: fakePVGetter{pvList: []byte(`No resources found.
{
    "apiVersion": "v1",
    "items": [],
    "kind": "List",
    "metadata": {},
    "resourceVersion": "",
    "selfLink": ""
}`)},
			podGetter: fakePodGetter{podList: []byte(`No resources found.
{
    "apiVersion": "v1",
    "items": [],
    "kind": "List",
    "metadata": {},
    "resourceVersion": "",
    "selfLink": ""
}`)},
			shouldBeNil:  false,
			shouldError:  false,
			volumesCount: 1,
		},
		volumeListTester{
			glusterGetter: fakeGlusterGetter{glusterVolumeList: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cliOutput>
  <opRet>0</opRet>
  <opErrno>0</opErrno>
  <opErrstr/>
  <volInfo>
    <volumes>
      <volume>
        <name>storage1</name>
        <id>f7803d45-f974-4832-81d1-1aa9db4be522</id>
        <status>1</status>
        <statusStr>Started</statusStr>
        <snapshotCount>0</snapshotCount>
        <brickCount>1</brickCount>
        <distCount>1</distCount>
        <stripeCount>1</stripeCount>
        <replicaCount>1</replicaCount>
        <arbiterCount>0</arbiterCount>
        <disperseCount>0</disperseCount>
        <redundancyCount>0</redundancyCount>
        <type>0</type>
        <typeStr>Distribute</typeStr>
        <transport>0</transport>
        <xlators/>
        <bricks>
          <brick uuid="3cf478d7-27da-4382-8e9f-44cc72a7beb2">ip-10-0-3-199:/data/storage1<name>ip-10-0-3-199:/data/storage1</name><hostUuid>3cf478d7-27da-4382-8e9f-44cc72a7beb2</hostUuid><isArbiter>0</isArbiter></brick>
        </bricks>
        <optCount>7</optCount>
        <options>
          <option>
            <name>features.quota-deem-statfs</name>
            <value>on</value>
          </option>
          <option>
            <name>features.inode-quota</name>
            <value>on</value>
          </option>
          <option>
            <name>features.quota</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.rpc-auth-allow</name>
            <value>172.16.0.0/16,10.0.3.194,10.0.3.230,10.0.3.204,10.0.3.65,10.0.3.118,10.0.3.75,10.0.3.199</value>
          </option>
          <option>
            <name>transport.address-family</name>
            <value>inet</value>
          </option>
          <option>
            <name>performance.readdir-ahead</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.disable</name>
            <value>off</value>
          </option>
        </options>
      </volume>
      <volume>
        <name>storage2</name>
        <id>417c6c43-a5b8-44f2-ad8a-8b88ac6de61c</id>
        <status>1</status>
        <statusStr>Started</statusStr>
        <snapshotCount>0</snapshotCount>
        <brickCount>4</brickCount>
        <distCount>2</distCount>
        <stripeCount>1</stripeCount>
        <replicaCount>2</replicaCount>
        <arbiterCount>0</arbiterCount>
        <disperseCount>0</disperseCount>
        <redundancyCount>0</redundancyCount>
        <type>7</type>
        <typeStr>Distributed-Replicate</typeStr>
        <transport>0</transport>
        <xlators/>
        <bricks>
          <brick uuid="46de1325-c06c-4703-9360-6001e71dcda3">ip-10-0-3-65:/data/storage2<name>ip-10-0-3-65:/data/storage2</name><hostUuid>46de1325-c06c-4703-9360-6001e71dcda3</hostUuid><isArbiter>0</isArbiter></brick>
          <brick uuid="978b4f00-9eef-4100-9f15-939cd9dca6b0">ip-10-0-3-75:/data/storage2<name>ip-10-0-3-75:/data/storage2</name><hostUuid>978b4f00-9eef-4100-9f15-939cd9dca6b0</hostUuid><isArbiter>0</isArbiter></brick>
          <brick uuid="7ead4086-a33e-4f3d-99f3-a9f2d2f95178">ip-10-0-3-118:/data/storage2<name>ip-10-0-3-118:/data/storage2</name><hostUuid>7ead4086-a33e-4f3d-99f3-a9f2d2f95178</hostUuid><isArbiter>0</isArbiter></brick>
          <brick uuid="3cf478d7-27da-4382-8e9f-44cc72a7beb2">ip-10-0-3-199:/data/storage2<name>ip-10-0-3-199:/data/storage2</name><hostUuid>3cf478d7-27da-4382-8e9f-44cc72a7beb2</hostUuid><isArbiter>0</isArbiter></brick>
        </bricks>
        <optCount>7</optCount>
        <options>
          <option>
            <name>features.quota-deem-statfs</name>
            <value>on</value>
          </option>
          <option>
            <name>features.inode-quota</name>
            <value>on</value>
          </option>
          <option>
            <name>features.quota</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.rpc-auth-allow</name>
            <value>172.16.0.0/16,10.0.3.194,10.0.3.230,10.0.3.204,10.0.3.65,10.0.3.118,10.0.3.75,10.0.3.199</value>
          </option>
          <option>
            <name>transport.address-family</name>
            <value>inet</value>
          </option>
          <option>
            <name>performance.readdir-ahead</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.disable</name>
            <value>off</value>
          </option>
        </options>
      </volume>
      <count>2</count>
    </volumes>
  </volInfo>
</cliOutput>`),
				glusterQuotas: map[string][]byte{"storage1": []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cliOutput>
  <opRet>0</opRet>
  <opErrno>0</opErrno>
  <opErrstr/>
  <volQuota>
    <limit>
      <path>/</path>
      <hard_limit>1073741824</hard_limit>
      <soft_limit_percent>80%</soft_limit_percent>
      <soft_limit_value>858993459</soft_limit_value>
      <used_space>0</used_space>
      <avail_space>1073741824</avail_space>
      <sl_exceeded>No</sl_exceeded>
      <hl_exceeded>No</hl_exceeded>
    </limit>
  </volQuota>
</cliOutput>`)},
			},
			pvGetter: fakePVGetter{pvList: []byte(`{
    "apiVersion": "v1",
    "items": [
        {
            "apiVersion": "v1",
            "kind": "PersistentVolume",
            "metadata": {
                "annotations": {
                    "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"PersistentVolume\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"storage2\",\"creationTimestamp\":null,\"annotations\":{\"volume.beta.kubernetes.io/storage-class\":\"kismatic\"}},\"spec\":{\"capacity\":{\"storage\":\"1Gi\"},\"nfs\":{\"server\":\"172.17.146.174\",\"path\":\"/storage2\"},\"accessModes\":[\"ReadWriteMany\"],\"persistentVolumeReclaimPolicy\":\"Retain\"},\"status\":{}}",
                    "volume.beta.kubernetes.io/storage-class": "kismatic"
                },
                "creationTimestamp": "2017-01-23T17:32:45Z",
                "name": "storage2",
                "namespace": "",
                "resourceVersion": "19526",
                "selfLink": "/api/v1/persistentvolumesstorage2",
                "uid": "f3d3846f-e191-11e6-a892-129f29c68938"
            },
            "spec": {
                "accessModes": [
                    "ReadWriteMany"
                ],
                "capacity": {
                    "storage": "1Gi"
                },
                "nfs": {
                    "path": "/storage2",
                    "server": "172.17.146.174"
                },
                "persistentVolumeReclaimPolicy": "Retain"
            },
            "status": {
                "phase": "Available"
            }
        }
    ],
    "kind": "List",
    "metadata": {},
    "resourceVersion": "",
    "selfLink": ""
}`)},
			podGetter: fakePodGetter{podList: []byte(`No resources found.
{
    "apiVersion": "v1",
    "items": [],
    "kind": "List",
    "metadata": {},
    "resourceVersion": "",
    "selfLink": ""
}`)},
			shouldBeNil:  false,
			shouldError:  false,
			volumesCount: 2,
		},
		volumeListTester{
			glusterGetter: fakeGlusterGetter{glusterVolumeList: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cliOutput>
  <opRet>0</opRet>
  <opErrno>0</opErrno>
  <opErrstr/>
  <volInfo>
    <volumes>
      <volume>
        <name>storage1</name>
        <id>f7803d45-f974-4832-81d1-1aa9db4be522</id>
        <status>1</status>
        <statusStr>Started</statusStr>
        <snapshotCount>0</snapshotCount>
        <brickCount>1</brickCount>
        <distCount>1</distCount>
        <stripeCount>1</stripeCount>
        <replicaCount>1</replicaCount>
        <arbiterCount>0</arbiterCount>
        <disperseCount>0</disperseCount>
        <redundancyCount>0</redundancyCount>
        <type>0</type>
        <typeStr>Distribute</typeStr>
        <transport>0</transport>
        <xlators/>
        <bricks>
          <brick uuid="3cf478d7-27da-4382-8e9f-44cc72a7beb2">ip-10-0-3-199:/data/storage1<name>ip-10-0-3-199:/data/storage1</name><hostUuid>3cf478d7-27da-4382-8e9f-44cc72a7beb2</hostUuid><isArbiter>0</isArbiter></brick>
        </bricks>
        <optCount>7</optCount>
        <options>
          <option>
            <name>features.quota-deem-statfs</name>
            <value>on</value>
          </option>
          <option>
            <name>features.inode-quota</name>
            <value>on</value>
          </option>
          <option>
            <name>features.quota</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.rpc-auth-allow</name>
            <value>172.16.0.0/16,10.0.3.194,10.0.3.230,10.0.3.204,10.0.3.65,10.0.3.118,10.0.3.75,10.0.3.199</value>
          </option>
          <option>
            <name>transport.address-family</name>
            <value>inet</value>
          </option>
          <option>
            <name>performance.readdir-ahead</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.disable</name>
            <value>off</value>
          </option>
        </options>
      </volume>
      <volume>
        <name>storage2</name>
        <id>417c6c43-a5b8-44f2-ad8a-8b88ac6de61c</id>
        <status>1</status>
        <statusStr>Started</statusStr>
        <snapshotCount>0</snapshotCount>
        <brickCount>4</brickCount>
        <distCount>2</distCount>
        <stripeCount>1</stripeCount>
        <replicaCount>2</replicaCount>
        <arbiterCount>0</arbiterCount>
        <disperseCount>0</disperseCount>
        <redundancyCount>0</redundancyCount>
        <type>7</type>
        <typeStr>Distributed-Replicate</typeStr>
        <transport>0</transport>
        <xlators/>
        <bricks>
          <brick uuid="46de1325-c06c-4703-9360-6001e71dcda3">ip-10-0-3-65:/data/storage2<name>ip-10-0-3-65:/data/storage2</name><hostUuid>46de1325-c06c-4703-9360-6001e71dcda3</hostUuid><isArbiter>0</isArbiter></brick>
          <brick uuid="978b4f00-9eef-4100-9f15-939cd9dca6b0">ip-10-0-3-75:/data/storage2<name>ip-10-0-3-75:/data/storage2</name><hostUuid>978b4f00-9eef-4100-9f15-939cd9dca6b0</hostUuid><isArbiter>0</isArbiter></brick>
          <brick uuid="7ead4086-a33e-4f3d-99f3-a9f2d2f95178">ip-10-0-3-118:/data/storage2<name>ip-10-0-3-118:/data/storage2</name><hostUuid>7ead4086-a33e-4f3d-99f3-a9f2d2f95178</hostUuid><isArbiter>0</isArbiter></brick>
          <brick uuid="3cf478d7-27da-4382-8e9f-44cc72a7beb2">ip-10-0-3-199:/data/storage2<name>ip-10-0-3-199:/data/storage2</name><hostUuid>3cf478d7-27da-4382-8e9f-44cc72a7beb2</hostUuid><isArbiter>0</isArbiter></brick>
        </bricks>
        <optCount>7</optCount>
        <options>
          <option>
            <name>features.quota-deem-statfs</name>
            <value>on</value>
          </option>
          <option>
            <name>features.inode-quota</name>
            <value>on</value>
          </option>
          <option>
            <name>features.quota</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.rpc-auth-allow</name>
            <value>172.16.0.0/16,10.0.3.194,10.0.3.230,10.0.3.204,10.0.3.65,10.0.3.118,10.0.3.75,10.0.3.199</value>
          </option>
          <option>
            <name>transport.address-family</name>
            <value>inet</value>
          </option>
          <option>
            <name>performance.readdir-ahead</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.disable</name>
            <value>off</value>
          </option>
        </options>
      </volume>
      <count>2</count>
    </volumes>
  </volInfo>
</cliOutput>`),
				glusterQuotas: map[string][]byte{"storage1": []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cliOutput>
  <opRet>0</opRet>
  <opErrno>0</opErrno>
  <opErrstr/>
  <volQuota>
    <limit>
      <path>/</path>
      <hard_limit>1073741824</hard_limit>
      <soft_limit_percent>80%</soft_limit_percent>
      <soft_limit_value>858993459</soft_limit_value>
      <used_space>0</used_space>
      <avail_space>1073741824</avail_space>
      <sl_exceeded>No</sl_exceeded>
      <hl_exceeded>No</hl_exceeded>
    </limit>
  </volQuota>
</cliOutput>`)},
			},
			pvGetter: fakePVGetter{pvList: []byte(`{
    "apiVersion": "v1",
    "items": [
        {
            "apiVersion": "v1",
            "kind": "PersistentVolume",
            "metadata": {
                "annotations": {
                    "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"PersistentVolume\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"storage2\",\"creationTimestamp\":null,\"annotations\":{\"volume.beta.kubernetes.io/storage-class\":\"kismatic\"}},\"spec\":{\"capacity\":{\"storage\":\"1Gi\"},\"nfs\":{\"server\":\"172.17.146.174\",\"path\":\"/storage2\"},\"accessModes\":[\"ReadWriteMany\"],\"persistentVolumeReclaimPolicy\":\"Retain\"},\"status\":{}}",
                    "pv.kubernetes.io/bound-by-controller": "yes",
                    "volume.beta.kubernetes.io/storage-class": "kismatic"
                },
                "creationTimestamp": "2017-01-23T17:32:45Z",
                "name": "storage2",
                "namespace": "",
                "resourceVersion": "20543",
                "selfLink": "/api/v1/persistentvolumesstorage2",
                "uid": "f3d3846f-e191-11e6-a892-129f29c68938"
            },
            "spec": {
                "accessModes": [
                    "ReadWriteMany"
                ],
                "capacity": {
                    "storage": "1Gi"
                },
                "claimRef": {
                    "apiVersion": "v1",
                    "kind": "PersistentVolumeClaim",
                    "name": "kismatic-integration-claim",
                    "namespace": "default",
                    "resourceVersion": "20541",
                    "uid": "5786857e-e193-11e6-a892-129f29c68938"
                },
                "nfs": {
                    "path": "/storage2",
                    "server": "172.17.146.174"
                },
                "persistentVolumeReclaimPolicy": "Retain"
            },
            "status": {
                "phase": "Bound"
            }
        }
    ],
    "kind": "List",
    "metadata": {},
    "resourceVersion": "",
    "selfLink": ""
}`)},
			podGetter: fakePodGetter{podList: []byte(`No resources found.
{
    "apiVersion": "v1",
    "items": [],
    "kind": "List",
    "metadata": {},
    "resourceVersion": "",
    "selfLink": ""
}`)},
			shouldBeNil:      false,
			shouldError:      false,
			boundVolumeCount: 1,
			volumesCount:     2,
		},
		volumeListTester{
			glusterGetter: fakeGlusterGetter{glusterVolumeList: []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cliOutput>
  <opRet>0</opRet>
  <opErrno>0</opErrno>
  <opErrstr/>
  <volInfo>
    <volumes>
      <volume>
        <name>storage1</name>
        <id>f7803d45-f974-4832-81d1-1aa9db4be522</id>
        <status>1</status>
        <statusStr>Started</statusStr>
        <snapshotCount>0</snapshotCount>
        <brickCount>1</brickCount>
        <distCount>1</distCount>
        <stripeCount>1</stripeCount>
        <replicaCount>1</replicaCount>
        <arbiterCount>0</arbiterCount>
        <disperseCount>0</disperseCount>
        <redundancyCount>0</redundancyCount>
        <type>0</type>
        <typeStr>Distribute</typeStr>
        <transport>0</transport>
        <xlators/>
        <bricks>
          <brick uuid="3cf478d7-27da-4382-8e9f-44cc72a7beb2">ip-10-0-3-199:/data/storage1<name>ip-10-0-3-199:/data/storage1</name><hostUuid>3cf478d7-27da-4382-8e9f-44cc72a7beb2</hostUuid><isArbiter>0</isArbiter></brick>
        </bricks>
        <optCount>7</optCount>
        <options>
          <option>
            <name>features.quota-deem-statfs</name>
            <value>on</value>
          </option>
          <option>
            <name>features.inode-quota</name>
            <value>on</value>
          </option>
          <option>
            <name>features.quota</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.rpc-auth-allow</name>
            <value>172.16.0.0/16,10.0.3.194,10.0.3.230,10.0.3.204,10.0.3.65,10.0.3.118,10.0.3.75,10.0.3.199</value>
          </option>
          <option>
            <name>transport.address-family</name>
            <value>inet</value>
          </option>
          <option>
            <name>performance.readdir-ahead</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.disable</name>
            <value>off</value>
          </option>
        </options>
      </volume>
      <volume>
        <name>storage2</name>
        <id>417c6c43-a5b8-44f2-ad8a-8b88ac6de61c</id>
        <status>1</status>
        <statusStr>Started</statusStr>
        <snapshotCount>0</snapshotCount>
        <brickCount>4</brickCount>
        <distCount>2</distCount>
        <stripeCount>1</stripeCount>
        <replicaCount>2</replicaCount>
        <arbiterCount>0</arbiterCount>
        <disperseCount>0</disperseCount>
        <redundancyCount>0</redundancyCount>
        <type>7</type>
        <typeStr>Distributed-Replicate</typeStr>
        <transport>0</transport>
        <xlators/>
        <bricks>
          <brick uuid="46de1325-c06c-4703-9360-6001e71dcda3">ip-10-0-3-65:/data/storage2<name>ip-10-0-3-65:/data/storage2</name><hostUuid>46de1325-c06c-4703-9360-6001e71dcda3</hostUuid><isArbiter>0</isArbiter></brick>
          <brick uuid="978b4f00-9eef-4100-9f15-939cd9dca6b0">ip-10-0-3-75:/data/storage2<name>ip-10-0-3-75:/data/storage2</name><hostUuid>978b4f00-9eef-4100-9f15-939cd9dca6b0</hostUuid><isArbiter>0</isArbiter></brick>
          <brick uuid="7ead4086-a33e-4f3d-99f3-a9f2d2f95178">ip-10-0-3-118:/data/storage2<name>ip-10-0-3-118:/data/storage2</name><hostUuid>7ead4086-a33e-4f3d-99f3-a9f2d2f95178</hostUuid><isArbiter>0</isArbiter></brick>
          <brick uuid="3cf478d7-27da-4382-8e9f-44cc72a7beb2">ip-10-0-3-199:/data/storage2<name>ip-10-0-3-199:/data/storage2</name><hostUuid>3cf478d7-27da-4382-8e9f-44cc72a7beb2</hostUuid><isArbiter>0</isArbiter></brick>
        </bricks>
        <optCount>7</optCount>
        <options>
          <option>
            <name>features.quota-deem-statfs</name>
            <value>on</value>
          </option>
          <option>
            <name>features.inode-quota</name>
            <value>on</value>
          </option>
          <option>
            <name>features.quota</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.rpc-auth-allow</name>
            <value>172.16.0.0/16,10.0.3.194,10.0.3.230,10.0.3.204,10.0.3.65,10.0.3.118,10.0.3.75,10.0.3.199</value>
          </option>
          <option>
            <name>transport.address-family</name>
            <value>inet</value>
          </option>
          <option>
            <name>performance.readdir-ahead</name>
            <value>on</value>
          </option>
          <option>
            <name>nfs.disable</name>
            <value>off</value>
          </option>
        </options>
      </volume>
      <count>2</count>
    </volumes>
  </volInfo>
</cliOutput>`),
				glusterQuotas: map[string][]byte{"storage1": []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cliOutput>
  <opRet>0</opRet>
  <opErrno>0</opErrno>
  <opErrstr/>
  <volQuota>
    <limit>
      <path>/</path>
      <hard_limit>1073741824</hard_limit>
      <soft_limit_percent>80%</soft_limit_percent>
      <soft_limit_value>858993459</soft_limit_value>
      <used_space>0</used_space>
      <avail_space>1073741824</avail_space>
      <sl_exceeded>No</sl_exceeded>
      <hl_exceeded>No</hl_exceeded>
    </limit>
  </volQuota>
</cliOutput>`)},
			},
			pvGetter: fakePVGetter{pvList: []byte(`{
    "apiVersion": "v1",
    "items": [
        {
            "apiVersion": "v1",
            "kind": "PersistentVolume",
            "metadata": {
                "annotations": {
                    "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"PersistentVolume\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"storage2\",\"creationTimestamp\":null,\"annotations\":{\"volume.beta.kubernetes.io/storage-class\":\"kismatic\"}},\"spec\":{\"capacity\":{\"storage\":\"1Gi\"},\"nfs\":{\"server\":\"172.17.146.174\",\"path\":\"/storage2\"},\"accessModes\":[\"ReadWriteMany\"],\"persistentVolumeReclaimPolicy\":\"Retain\"},\"status\":{}}",
                    "pv.kubernetes.io/bound-by-controller": "yes",
                    "volume.beta.kubernetes.io/storage-class": "kismatic"
                },
                "creationTimestamp": "2017-01-23T17:32:45Z",
                "name": "storage2",
                "namespace": "",
                "resourceVersion": "20543",
                "selfLink": "/api/v1/persistentvolumesstorage2",
                "uid": "f3d3846f-e191-11e6-a892-129f29c68938"
            },
            "spec": {
                "accessModes": [
                    "ReadWriteMany"
                ],
                "capacity": {
                    "storage": "1Gi"
                },
                "claimRef": {
                    "apiVersion": "v1",
                    "kind": "PersistentVolumeClaim",
                    "name": "kismatic-integration-claim",
                    "namespace": "default",
                    "resourceVersion": "20541",
                    "uid": "5786857e-e193-11e6-a892-129f29c68938"
                },
                "nfs": {
                    "path": "/storage2",
                    "server": "172.17.146.174"
                },
                "persistentVolumeReclaimPolicy": "Retain"
            },
            "status": {
                "phase": "Bound"
            }
        }
    ],
    "kind": "List",
    "metadata": {},
    "resourceVersion": "",
    "selfLink": ""
}`)},
			podGetter: fakePodGetter{podList: []byte(`{
    "apiVersion": "v1",
    "items": [
        {
            "apiVersion": "v1",
            "kind": "Pod",
            "metadata": {
                "annotations": {
                    "kubectl.kubernetes.io/last-applied-configuration": "{\"kind\":\"Pod\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"mypod\",\"creationTimestamp\":null},\"spec\":{\"volumes\":[{\"name\":\"mypd\",\"persistentVolumeClaim\":{\"claimName\":\"kismatic-integration-claim\"}}],\"containers\":[{\"name\":\"myfrontend\",\"image\":\"nginx\",\"resources\":{},\"volumeMounts\":[{\"name\":\"mypd\",\"mountPath\":\"/var/www/html\"}]}]},\"status\":{}}"
                },
                "creationTimestamp": "2017-01-23T17:49:32Z",
                "name": "mypod",
                "namespace": "default",
                "resourceVersion": "21288",
                "selfLink": "/api/v1/namespaces/default/pods/mypod",
                "uid": "4b9ffa36-e194-11e6-a892-129f29c68938"
            },
            "spec": {
                "containers": [
                    {
                        "image": "nginx",
                        "imagePullPolicy": "Always",
                        "name": "myfrontend",
                        "resources": {},
                        "terminationMessagePath": "/dev/termination-log",
                        "volumeMounts": [
                            {
                                "mountPath": "/var/www/html",
                                "name": "mypd"
                            },
                            {
                                "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
                                "name": "default-token-03rfm",
                                "readOnly": true
                            }
                        ]
                    }
                ],
                "dnsPolicy": "ClusterFirst",
                "nodeName": "ip-10-0-3-230",
                "restartPolicy": "Always",
                "securityContext": {},
                "serviceAccount": "default",
                "serviceAccountName": "default",
                "terminationGracePeriodSeconds": 30,
                "volumes": [
                    {
                        "name": "mypd",
                        "persistentVolumeClaim": {
                            "claimName": "kismatic-integration-claim"
                        }
                    },
                    {
                        "name": "default-token-03rfm",
                        "secret": {
                            "defaultMode": 420,
                            "secretName": "default-token-03rfm"
                        }
                    }
                ]
            },
            "status": {
                "conditions": [
                    {
                        "lastProbeTime": null,
                        "lastTransitionTime": "2017-01-23T17:49:32Z",
                        "status": "True",
                        "type": "Initialized"
                    },
                    {
                        "lastProbeTime": null,
                        "lastTransitionTime": "2017-01-23T17:49:57Z",
                        "status": "True",
                        "type": "Ready"
                    },
                    {
                        "lastProbeTime": null,
                        "lastTransitionTime": "2017-01-23T17:49:32Z",
                        "status": "True",
                        "type": "PodScheduled"
                    }
                ],
                "containerStatuses": [
                    {
                        "containerID": "docker://62f01ca6580f6b5f9fecd841e2450e3f71dec07c3a6b867d95627baa3dd6a475",
                        "image": "nginx",
                        "imageID": "docker://sha256:a39777a1a4a6ec8a91c978ded905cca10e6b105ba650040e16c50b3e157272c3",
                        "lastState": {},
                        "name": "myfrontend",
                        "ready": true,
                        "restartCount": 0,
                        "state": {
                            "running": {
                                "startedAt": "2017-01-23T17:49:56Z"
                            }
                        }
                    }
                ],
                "hostIP": "10.0.3.230",
                "phase": "Running",
                "podIP": "172.16.255.135",
                "startTime": "2017-01-23T17:49:32Z"
            }
        },
        {
            "apiVersion": "v1",
            "kind": "Pod",
            "metadata": {
                "annotations": {
                    "kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\",\"namespace\":\"kube-system\",\"name\":\"kubernetes-dashboard-1280404318\",\"uid\":\"18b60843-e178-11e6-a892-129f29c68938\",\"apiVersion\":\"extensions\",\"resourceVersion\":\"371\"}}\n"
                },
                "creationTimestamp": "2017-01-23T14:27:40Z",
                "generateName": "kubernetes-dashboard-1280404318-",
                "labels": {
                    "app": "kubernetes-dashboard",
                    "pod-template-hash": "1280404318"
                },
                "name": "kubernetes-dashboard-1280404318-n5mqh",
                "namespace": "kube-system",
                "ownerReferences": [
                    {
                        "apiVersion": "extensions/v1beta1",
                        "controller": true,
                        "kind": "ReplicaSet",
                        "name": "kubernetes-dashboard-1280404318",
                        "uid": "18b60843-e178-11e6-a892-129f29c68938"
                    }
                ],
                "resourceVersion": "466",
                "selfLink": "/api/v1/namespaces/kube-system/pods/kubernetes-dashboard-1280404318-n5mqh",
                "uid": "18b70e5a-e178-11e6-a892-129f29c68938"
            },
            "spec": {
                "containers": [
                    {
                        "image": "gcr.io/google_containers/kubernetes-dashboard-amd64:v1.5.0",
                        "imagePullPolicy": "IfNotPresent",
                        "livenessProbe": {
                            "failureThreshold": 3,
                            "httpGet": {
                                "path": "/",
                                "port": 9090,
                                "scheme": "HTTP"
                            },
                            "initialDelaySeconds": 30,
                            "periodSeconds": 10,
                            "successThreshold": 1,
                            "timeoutSeconds": 30
                        },
                        "name": "kubernetes-dashboard",
                        "ports": [
                            {
                                "containerPort": 9090,
                                "protocol": "TCP"
                            }
                        ],
                        "resources": {},
                        "terminationMessagePath": "/dev/termination-log",
                        "volumeMounts": [
                            {
                                "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
                                "name": "default-token-gn2nc",
                                "readOnly": true
                            }
                        ]
                    }
                ],
                "dnsPolicy": "ClusterFirst",
                "nodeName": "ip-10-0-3-230",
                "restartPolicy": "Always",
                "securityContext": {},
                "serviceAccount": "default",
                "serviceAccountName": "default",
                "terminationGracePeriodSeconds": 30,
                "volumes": [
                    {
                        "name": "default-token-gn2nc",
                        "secret": {
                            "defaultMode": 420,
                            "secretName": "default-token-gn2nc"
                        }
                    }
                ]
            },
            "status": {
                "conditions": [
                    {
                        "lastProbeTime": null,
                        "lastTransitionTime": "2017-01-23T14:27:40Z",
                        "status": "True",
                        "type": "Initialized"
                    },
                    {
                        "lastProbeTime": null,
                        "lastTransitionTime": "2017-01-23T14:28:03Z",
                        "status": "True",
                        "type": "Ready"
                    },
                    {
                        "lastProbeTime": null,
                        "lastTransitionTime": "2017-01-23T14:27:40Z",
                        "status": "True",
                        "type": "PodScheduled"
                    }
                ],
                "containerStatuses": [
                    {
                        "containerID": "docker://78cd2bfd9ea6750c1cfeddef0cfd53466586e71d583c0928985e6001e01a0141",
                        "image": "gcr.io/google_containers/kubernetes-dashboard-amd64:v1.5.0",
                        "imageID": "docker://sha256:e5133bac8024ac6c916f16df8790259b5504a800766bee87dcf90ec7d634a418",
                        "lastState": {},
                        "name": "kubernetes-dashboard",
                        "ready": true,
                        "restartCount": 0,
                        "state": {
                            "running": {
                                "startedAt": "2017-01-23T14:28:02Z"
                            }
                        }
                    }
                ],
                "hostIP": "10.0.3.230",
                "phase": "Running",
                "podIP": "172.16.255.130",
                "startTime": "2017-01-23T14:27:40Z"
            }
        }
    ],
    "kind": "List",
    "metadata": {},
    "resourceVersion": "",
    "selfLink": ""
}`)},
			shouldBeNil:        false,
			shouldError:        false,
			boundVolumeCount:   1,
			claimedVolumeCount: 1,
			volumesCount:       2,
		},
	}

	for _, test := range tests {
		resp, err := buildResponse(test.glusterGetter, test.pvGetter, test.podGetter)
		if err != nil && !test.shouldError {
			t.Errorf("unexpected error: %v, for %v", err, test)
		}
		if err == nil && test.shouldError {
			t.Errorf("expected an error but got nil for %v", test)
		}
		if resp == nil && !test.shouldBeNil {
			t.Errorf("did not expect response to be nil, error: %v", err)
		}
		if resp != nil && test.shouldBeNil {
			t.Errorf("expected response to be nil but got %v", resp)
		}
		if !test.shouldError && !test.shouldBeNil && resp != nil {
			if len(resp.Volumes) != test.volumesCount {
				t.Errorf("expected to get %d volumes, instead got %d, %v", test.volumesCount, len(resp.Volumes), resp)
			}
			var bound, claimed int
			for _, v := range resp.Volumes {
				if v.Claim != nil {
					bound = bound + 1
				}
				if v.Pods != nil && len(v.Pods) > 0 {
					claimed = claimed + 1
				}
			}
			if bound != test.boundVolumeCount {
				t.Errorf("expected to get %d bound volumes, instead got %d, %v", test.boundVolumeCount, bound, resp)
			}
			if claimed != test.claimedVolumeCount {
				t.Errorf("expected to get %d claimed volumes, instead got %d, %v", test.claimedVolumeCount, claimed, resp)
			}
		}
	}
}
