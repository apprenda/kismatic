package data

type PodList struct {
	Items []Pod `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type Pod struct {
	// Standard object's metadata.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	// +optional
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Specification of the desired behavior of the pod.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
	// +optional
	Spec PodSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}
type ObjectMeta struct {
	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`

	// Name must be unique within a namespace. Is required when creating resources, although
	// some resources may allow a client to request the generation of an appropriate name
	// automatically. Name is primarily intended for creation idempotence and configuration
	// definition.
	// Cannot be updated.
	// More info: http://kubernetes.io/docs/user-guide/identifiers#names
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// Namespace defines the space within each name must be unique. An empty namespace is
	// equivalent to the "default" namespace, but "default" is the canonical representation.
	// Not all objects are required to be scoped to a namespace - the value of this field for
	// those objects will be empty.
	//
	// Must be a DNS_LABEL.
	// Cannot be updated.
	// More info: http://kubernetes.io/docs/user-guide/namespaces
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://kubernetes.io/docs/user-guide/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`
}

// PodSpec is a description of a pod.
type PodSpec struct {
	// List of volumes that can be mounted by containers belonging to the pod.
	// More info: http://kubernetes.io/docs/user-guide/volumes
	// +optional
	Volumes []Volume `json:"volumes,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,1,rep,name=volumes"`
	// List of containers belonging to the pod.
	// Containers cannot currently be added or removed.
	// There must be at least one container in a Pod.
	// Cannot be updated.
	// More info: http://kubernetes.io/docs/user-guide/containers
	Containers []Container `json:"containers" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`
}

// Volume represents a named volume in a pod that may be accessed by any container in the pod.
type Volume struct {
	// Volume's name.
	// Must be a DNS_LABEL and unique within the pod.
	// More info: http://kubernetes.io/docs/user-guide/identifiers#names
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// VolumeSource represents the location and type of the mounted volume.
	// If not specified, the Volume is implied to be an EmptyDir.
	// This implied behavior is deprecated and will be removed in a future version.
	VolumeSource `json:",inline" protobuf:"bytes,2,opt,name=volumeSource"`
}

// Represents the source of a volume to mount.
// Only one of its members may be specified.
type VolumeSource struct {
	// PersistentVolumeClaimVolumeSource represents a reference to a
	// PersistentVolumeClaim in the same namespace.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#persistentvolumeclaims
	// +optional
	PersistentVolumeClaim *PersistentVolumeClaimVolumeSource `json:"persistentVolumeClaim,omitempty" protobuf:"bytes,10,opt,name=persistentVolumeClaim"`
}

// PersistentVolumeClaimVolumeSource references the user's PVC in the same namespace.
// This volume finds the bound PV and mounts that volume for the pod. A
// PersistentVolumeClaimVolumeSource is, essentially, a wrapper around another
// type of volume that is owned by someone else (the system).
type PersistentVolumeClaimVolumeSource struct {
	// ClaimName is the name of a PersistentVolumeClaim in the same namespace as the pod using this volume.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#persistentvolumeclaims
	ClaimName string `json:"claimName" protobuf:"bytes,1,opt,name=claimName"`
	// Will force the ReadOnly setting in VolumeMounts.
	// Default false.
	// +optional
	ReadOnly bool `json:"readOnly,omitempty" protobuf:"varint,2,opt,name=readOnly"`
}

// A single application container that you want to run within a pod.
type Container struct {
	// Name of the container specified as a DNS_LABEL.
	// Each container in a pod must have a unique name (DNS_LABEL).
	// Cannot be updated.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Pod volumes to mount into the container's filesystem.
	// Cannot be updated.
	// +optional
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty" patchStrategy:"merge" patchMergeKey:"mountPath" protobuf:"bytes,9,rep,name=volumeMounts"`
}

// VolumeMount describes a mounting of a Volume within a container.
type VolumeMount struct {
	// This must match the Name of a Volume.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Mounted read-only if true, read-write otherwise (false or unspecified).
	// Defaults to false.
	// +optional
	ReadOnly bool `json:"readOnly,omitempty" protobuf:"varint,2,opt,name=readOnly"`
	// Path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	MountPath string `json:"mountPath" protobuf:"bytes,3,opt,name=mountPath"`
	// Path within the volume from which the container's volume should be mounted.
	// Defaults to "" (volume's root).
	// +optional
	SubPath string `json:"subPath,omitempty" protobuf:"bytes,4,opt,name=subPath"`
}

// PersistentVolumeList is a list of PersistentVolume items.
type PersistentVolumeList struct {
	TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#types-kinds
	// +optional
	ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// List of persistent volumes.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes
	Items []PersistentVolume `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// PersistentVolume (PV) is a storage resource provisioned by an administrator.
// It is analogous to a node.
// More info: http://kubernetes.io/docs/user-guide/persistent-volumes
type PersistentVolume struct {
	TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	// +optional
	ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines a specification of a persistent volume owned by the cluster.
	// Provisioned by an administrator.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#persistent-volumes
	// +optional
	Spec PersistentVolumeSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status represents the current information/status for the persistent volume.
	// Populated by the system.
	// Read-only.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#persistent-volumes
	// +optional
	Status PersistentVolumeStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// TypeMeta describes an individual object in an API response or request
// with strings representing the type of the object and its API schema version.
// Structures that are versioned or persisted should inline TypeMeta.
type TypeMeta struct {
	// Kind is a string value representing the REST resource this object represents.
	// Servers may infer this from the endpoint the client submits requests to.
	// Cannot be updated.
	// In CamelCase.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#types-kinds
	// +optional
	Kind string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`

	// APIVersion defines the versioned schema of this representation of an object.
	// Servers should convert recognized schemas to the latest internal value, and
	// may reject unrecognized values.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#resources
	// +optional
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,2,opt,name=apiVersion"`
}

// ListMeta describes metadata that synthetic resources must have, including lists and
// various status objects. A resource may have only one of {ObjectMeta, ListMeta}.
type ListMeta struct {
	// SelfLink is a URL representing this object.
	// Populated by the system.
	// Read-only.
	// +optional
	SelfLink string `json:"selfLink,omitempty" protobuf:"bytes,1,opt,name=selfLink"`

	// String that identifies the server's internal version of this object that
	// can be used by clients to determine when objects have changed.
	// Value must be treated as opaque by clients and passed unmodified back to the server.
	// Populated by the system.
	// Read-only.
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#concurrency-control-and-consistency
	// +optional
	ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,2,opt,name=resourceVersion"`
}

// PersistentVolumeSpec is the specification of a persistent volume.
type PersistentVolumeSpec struct {
	// ClaimRef is part of a bi-directional binding between PersistentVolume and PersistentVolumeClaim.
	// Expected to be non-nil when bound.
	// claim.VolumeName is the authoritative bind between PV and PVC.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#binding
	// +optional
	ClaimRef *ObjectReference `json:"claimRef,omitempty" protobuf:"bytes,4,opt,name=claimRef"`
}

type PersistentVolumePhase string

// PersistentVolumeStatus is the current status of a persistent volume.
type PersistentVolumeStatus struct {
	// Phase indicates if a volume is available, bound to a claim, or released by a claim.
	// More info: http://kubernetes.io/docs/user-guide/persistent-volumes#phase
	// +optional
	Phase PersistentVolumePhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=PersistentVolumePhase"`
}

// ObjectReference contains enough information to let you inspect or modify the referred object.
type ObjectReference struct {
	// Namespace of the referent.
	// More info: http://kubernetes.io/docs/user-guide/namespaces
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
	// Name of the referent.
	// More info: http://kubernetes.io/docs/user-guide/identifiers#names
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,3,opt,name=name"`
}
