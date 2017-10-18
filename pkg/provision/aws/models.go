package aws

import "github.com/aws/aws-sdk-go/service/ec2"

type NodeBlueprint struct {
	EtcdInstanceType   InstanceType
	EtcdDisk           int64
	MasterInstanceType InstanceType
	MasterDisk         int64
	WorkerInstanceType InstanceType
	WorkerDisk         int64
}

var minimumMachine = NodeBlueprint{
	EtcdInstanceType:   ec2.InstanceTypeT2Micro,
	EtcdDisk:           12,
	MasterInstanceType: ec2.InstanceTypeT2Micro,
	MasterDisk:         12,
	WorkerInstanceType: ec2.InstanceTypeT2Micro,
	WorkerDisk:         12,
}

func newBlueprint(ms NodeBlueprint) NodeBlueprint {
	newDisk := minimumMachine
	if ms.EtcdInstanceType != "" {
		newDisk.EtcdInstanceType = ms.EtcdInstanceType
	}
	if ms.EtcdDisk > newDisk.EtcdDisk {
		newDisk.EtcdDisk = ms.EtcdDisk
	}

	if ms.MasterInstanceType != "" {
		newDisk.MasterInstanceType = ms.MasterInstanceType
	}
	if ms.MasterDisk > newDisk.MasterDisk {
		newDisk.MasterDisk = ms.MasterDisk
	}

	if ms.WorkerInstanceType != "" {
		newDisk.WorkerInstanceType = ms.WorkerInstanceType
	}
	if ms.WorkerDisk > newDisk.WorkerDisk {
		newDisk.WorkerDisk = ms.WorkerDisk
	}

	return newDisk
}

var (
	NodeBlueprintMap = make(map[string]NodeBlueprint)
)

func init() {
	NodeBlueprintMap["micro"] = newBlueprint(NodeBlueprint{})
	NodeBlueprintMap["small"] = newBlueprint(NodeBlueprint{
		WorkerInstanceType: ec2.InstanceTypeT2Medium,
	})
	NodeBlueprintMap["beefy"] = newBlueprint(NodeBlueprint{
		EtcdInstanceType:   ec2.InstanceTypeM4Large,
		EtcdDisk:           50,
		MasterInstanceType: ec2.InstanceTypeM4Xlarge,
		MasterDisk:         50,
		WorkerInstanceType: ec2.InstanceTypeM4Xlarge,
		WorkerDisk:         200,
	})
}
