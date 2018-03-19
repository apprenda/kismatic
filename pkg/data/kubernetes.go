package data

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apprenda/kismatic/pkg/ssh"
)

// PodLister lists pods on a Kubernetes cluster
type PodLister interface {
	ListPods() (*PodList, error)
}

// PVLister lists persistent volumes that exist on a Kubernetes cluster
type PVLister interface {
	ListPersistentVolumes() (*PersistentVolumeList, error)
}

// PersistentVolumeGetter gets a persistent volume
type PersistentVolumeGetter interface {
	GetPersistentVolume(name string) (*PersistentVolume, error)
}

// PersistentVolumeClaimGetter gets a persistent volume claim
type PersistentVolumeClaimGetter interface {
	GetPersistentVolumeClaim(namespace, name string) (*PersistentVolumeClaim, error)
}

// DaemonSetGetter gets a given daemonset
type DaemonSetGetter interface {
	GetDaemonSet(namespace, name string) (*DaemonSet, error)
}

// ReplicationControllerGetter gets a replication controller
type ReplicationControllerGetter interface {
	GetReplicationController(namespace, name string) (*ReplicationController, error)
}

// ReplicaSetGetter gets a replica set
type ReplicaSetGetter interface {
	GetReplicaSet(namespace, name string) (*ReplicaSet, error)
}

// StatefulSetGetter gets a stateful set
type StatefulSetGetter interface {
	GetStatefulSet(namespace, name string) (*StatefulSet, error)
}

type KubernetesClient interface {
	PodLister
	PVLister
}

// RemoteKubectl is a kubectl client that uses an underlying SSH connection
// to connect to a node that has the kubectl binary. It is expected that this
// node has access to a kubernetes cluster via kubectl.
type RemoteKubectl struct {
	SSHClient ssh.Client
}

// ListPersistentVolumes returns PersistentVolume data
func (k RemoteKubectl) ListPersistentVolumes() (*PersistentVolumeList, error) {
	pvRaw, err := k.SSHClient.Output(true, "sudo kubectl --kubeconfig /root/.kube/config get pv -o json")
	if err != nil {
		return nil, fmt.Errorf("error getting persistent volume data: %v", err)
	}
	return UnmarshalPVs(pvRaw)
}

func UnmarshalPVs(raw string) (*PersistentVolumeList, error) {
	if isNoResourcesResponse(raw) {
		return nil, nil
	}
	var pvs PersistentVolumeList
	err := json.Unmarshal([]byte(raw), &pvs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling persistent volume data: %v", err)
	}
	return &pvs, nil
}

// ListPods returns Pods data with --all-namespaces=true flag
func (k RemoteKubectl) ListPods() (*PodList, error) {
	podsRaw, err := k.SSHClient.Output(true, "sudo kubectl --kubeconfig /root/.kube/config get pods --all-namespaces=true -o json")
	if err != nil {
		return nil, fmt.Errorf("error getting pod data: %v", err)
	}
	return UnmarshalPods(podsRaw)
}

func UnmarshalPods(raw string) (*PodList, error) {
	if isNoResourcesResponse(raw) {
		return nil, nil
	}
	var pods PodList
	err := json.Unmarshal([]byte(raw), &pods)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling pod data: %v", err)
	}
	return &pods, nil
}

// GetDaemonSet returns the DaemonSet with the given namespace and name. If not found,
// returns an error.
func (k RemoteKubectl) GetDaemonSet(namespace, name string) (*DaemonSet, error) {
	cmd := fmt.Sprintf("sudo kubectl --kubeconfig /root/.kube/config get ds --namespace=%s -o json %s", namespace, name)
	dsRaw, err := k.SSHClient.Output(true, cmd)
	if err != nil {
		return nil, fmt.Errorf("error getting daemon sets: %v", err)
	}
	if isNoResourcesResponse(dsRaw) {
		return nil, fmt.Errorf("DaemonSet %s/%s was not found", namespace, name)
	}
	var d DaemonSet
	if err := json.Unmarshal([]byte(dsRaw), &d); err != nil {
		return nil, fmt.Errorf("error unmarshalling daemonset: %v", err)
	}
	return &d, nil
}

// GetReplicationController returns the ReplicationController with the given name in the given namespace.
// If not found, returns an error.
func (k RemoteKubectl) GetReplicationController(namespace, name string) (*ReplicationController, error) {
	cmd := fmt.Sprintf("sudo kubectl --kubeconfig /root/.kube/config get replicationcontroller --namespace=%s -o json %s", namespace, name)
	rcRaw, err := k.SSHClient.Output(true, cmd)
	if err != nil {
		return nil, fmt.Errorf("error getting replication controller: %v", err)
	}
	if isNoResourcesResponse(rcRaw) {
		return nil, fmt.Errorf("ReplicationController %s/%s was not found", namespace, name)
	}
	var r ReplicationController
	if err := json.Unmarshal([]byte(rcRaw), &r); err != nil {
		return nil, fmt.Errorf("error unmarshalling replication controller: %v", err)
	}
	return &r, nil
}

// GetReplicaSet returns the ReplicaSet with the given name in the given namespace.
// If not found, returns an error.
func (k RemoteKubectl) GetReplicaSet(namespace, name string) (*ReplicaSet, error) {
	cmd := fmt.Sprintf("sudo kubectl --kubeconfig /root/.kube/config get replicaset --namespace=%s -o json %s", namespace, name)
	raw, err := k.SSHClient.Output(true, cmd)
	if err != nil {
		return nil, fmt.Errorf("error getting ReplicaSet: %v", err)
	}
	if isNoResourcesResponse(raw) {
		return nil, fmt.Errorf("ReplicaSet %s/%s was not found", namespace, name)
	}
	var r ReplicaSet
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return nil, fmt.Errorf("error unmarshalling ReplicaSet: %v", err)
	}
	return &r, nil
}

// GetPersistentVolume returns the persistent volume with the given name.
// If not found, returns an error.
func (k RemoteKubectl) GetPersistentVolume(name string) (*PersistentVolume, error) {
	cmd := fmt.Sprintf("sudo kubectl --kubeconfig /root/.kube/config get pv -o json %s", name)
	raw, err := k.SSHClient.Output(true, cmd)
	if err != nil {
		return nil, fmt.Errorf("error getting PersistentVolume: %v", err)
	}
	if isNoResourcesResponse(raw) {
		return nil, fmt.Errorf("PersistentVolume %s was not found", name)
	}
	var p PersistentVolume
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return nil, fmt.Errorf("error unmarshalling PersistentVolume: %v", err)
	}
	return &p, nil
}

// GetPersistentVolumeClaim returns the persistent volume claim with the given name and namespace.
// If not found, returns an error.
func (k RemoteKubectl) GetPersistentVolumeClaim(namespace, name string) (*PersistentVolumeClaim, error) {
	cmd := fmt.Sprintf("sudo kubectl --kubeconfig /root/.kube/config get pvc --namespace %s -o json %s", namespace, name)
	raw, err := k.SSHClient.Output(true, cmd)
	if err != nil {
		return nil, fmt.Errorf("error getting PersistentVolumeClaim: %v", err)
	}
	if isNoResourcesResponse(raw) {
		return nil, fmt.Errorf("PersistentVolumeClaim %s was not found", name)
	}
	var p PersistentVolumeClaim
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return nil, fmt.Errorf("error unmarshalling PersistentVolumeClaim: %v", err)
	}
	return &p, nil
}

// GetStatefulSet returns the stateful set with the given name in the given namespace.
// If not found, returns an error.
func (k RemoteKubectl) GetStatefulSet(namespace, name string) (*StatefulSet, error) {
	cmd := fmt.Sprintf("sudo kubectl --kubeconfig /root/.kube/config get statefulset --namespace %s -o json %s", namespace, name)
	raw, err := k.SSHClient.Output(true, cmd)
	if err != nil {
		return nil, fmt.Errorf("error getting StatefulSet: %v", err)
	}
	if isNoResourcesResponse(raw) {
		return nil, fmt.Errorf("StatefulSet %s/%s was not found", namespace, name)
	}
	var s StatefulSet
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil, fmt.Errorf("error unmarshalling StatefulSet: %v", err)
	}
	return &s, nil
}

// kubectl will print this message when no resources are returned
func isNoResourcesResponse(s string) bool {
	if strings.Contains(strings.TrimSpace(s), "No resources found") {
		return true
	}
	return false
}
