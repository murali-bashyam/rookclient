package v1

import (
	"net"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeInfo struct {
	hostName string
	ipaddr   net.IP
}

type DevClass string

const (
	DevStandard DevClass = "standard" // maps to rotational devices, HDD
	DevMedium   DevClass = "medium"   // maps to SSD
	DevFast     DevClass = "fast"     // maps to NVMe
)

type FailureDomain string

const (
	FailureDomainHost FailureDomain = "host"
	FailureDomainRack FailureDomain = "rack"
)

// This CRD defines the storage policy parameters related to Durability.
// DurabilityClass can be "replication" or "erasurecoded".
// replication trades off storage for CPU.
// erasurecoded trades CPU for storage.

// DurabilityLevel determines amount of redundancy.
// Higher levels use more storage.

// For DurabilityClass "replication":
// DurabilityLevel : low --> replication factor 1
// DurabilityLevel : semi --> replication factor 2
// DurabilityLevel : normal --> replication factor 3
// DurabilityLevel : high --> replication factor 4

// For DurabilityClass "erasurecoded":
// DurabilityLevel : semi --> dataChunks : 2  codingChunks: 1
// DurabilityLevel : normal --> dataChunks : 3 codingChunks: 2
// DurabilityLevel : high --> dataChunks : 4 codingChunks: 3

type DurabilityClass string

const (
	DurabilityClassReplicated   DurabilityClass = "replicated"
	DurabilityClassErasureCoded DurabilityClass = "erasurecoded"
)

type DurabilityLevel string

const (
	DurabilityLevelLow    DurabilityLevel = "low"
	DurabilityLevelSemi   DurabilityLevel = "semi"
	DurabilityLevelNormal DurabilityLevel = "normal"
	DurabilityLevelHigh   DurabilityLevel = "high"
)

type StoragePolicyDurability struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// This field specifies the failure domain of storage servers
	// making up the pool.
	// Can be one of "rack" or "host", extended later.
	FailureDomain   FailureDomain   `json:"failuredomain"`
	DurabilityClass DurabilityClass `json:"durabilityclass"`
	DurabilityLevel DurabilityLevel `json: "redundancylevel"`
}

// This policy specifies the level of storage performance desired.
// This setting determines the type of storage devices which should
// be considered to build the storage pool.
// Can be one of "standard" or "medium" or "fast".
// standard maps to rotational devices, HDD.
// medium maps to SSD.
// fast maps to NVMe.
type StoragePolicyPerformance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	IoPerfClass DevClass `json:"ioperfclass"`
}

// The storage cluster object
// The storage cluster object
// A CRD based on this object is used to create the storage cluster.
// A CRD based on this object is injected into the user cluster.

type StorageClusterSpec struct {
	// Cluster ID of the storage cluster
	// if consuming storage from an external cluster.
	StorageClusterID string `json:"storageclusterid, omitempty"`

	// List of nodes which should be included as part of storage
	// cluster, they are dedicated for storage. If unspecified, all nodes
	// of the cluster will be assumed dedicated for storage.
	Nodelist []NodeInfo `json:"nodelist,omitempty"`

	// Prometheus monitoring, enabled by default.
	Monitoring bool `json"monitoring,omitempty"`
}

type StorageClusterPhase string

const (
	ClusterPhaseIgnored     StorageClusterPhase = "Ignored"
	ClusterPhaseConnecting  StorageClusterPhase = "Connecting"
	ClusterPhaseConnected   StorageClusterPhase = "Connected"
	ClusterPhaseProgressing StorageClusterPhase = "Progressing"
	ClusterPhaseReady       StorageClusterPhase = "Ready"
	ClusterPhaseUpdating    StorageClusterPhase = "Updating"
	ClusterPhaseFailure     StorageClusterPhase = "Failure"
	ClusterPhaseUpgrading   StorageClusterPhase = "Upgrading"
	ClusterPhaseDeleting    StorageClusterPhase = "Deleting"
)

type StorageClusterState string

const (
	ClusterStateCreating   StorageClusterState = "Creating"
	ClusterStateCreated    StorageClusterState = "Created"
	ClusterStateUpdating   StorageClusterState = "Updating"
	ClusterStateConnecting StorageClusterState = "Connecting"
	ClusterStateConnected  StorageClusterState = "Connected"
	ClusterStateError      StorageClusterState = "Error"
)

type StorageClusterStatus struct {
	// State indicates state of cluster
	State StorageClusterState `json:"state,omitempty"`

	// Phase indicates stage of cluster operations
	Phase StorageClusterPhase `json:"phase,omitempty"`

	// Message provides an explanation of the cluster phase
	Message string `json:"message,omitempty"`
}

type StorageCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StorageClusterSpec   `json:"spec"`
	Status StorageClusterStatus `json:"status"`
}

// The storage pool object
// A CRD based on this object is injected into the storage cluster.

type StoragePoolSpec struct {
	// This field specifies the storage cluster ID.
	ClusterID string `json:"clusterid"`

	// This field specifies any quota to set on the pool.
	// if unspecified, default is to use all available capacity of the cluster.
	Quota uint64 `json:"quota, omitempty"`

	// This field specifies any durability policy to set on the pool.
	// if unspecified, default DurabilityClass is replicated,
	// DurabilityLevel is normal.
	DurabilityPolicy StoragePolicyDurability `json:"durabilitypolicy, omitempty"`

	// This field specifies the performance policy to set on the pool.
	// if unspecified, default Ioperfclass is use to all available raw devices.
	PerfPolicy StoragePolicyPerformance `json:"perfpolicy, omitempty"`
}

type StoragePoolPhase string

const (
	PoolPhaseConnecting StoragePoolPhase = "Connecting"
	PoolPhaseReady      StoragePoolPhase = "Ready"
	PoolPhaseFailure    StoragePoolPhase = "Failure"
	PoolPhaseDeleting   StoragePoolPhase = "Deleting"
)

type StoragePoolStatus struct {
	// Phase indicates state of pool creation or deletion
	Phase StoragePoolPhase `json:"phase,omitempty"`
}

type StoragePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StoragePoolSpec   `json:"spec"`
	Status StoragePoolStatus `json:"status"`
}

type VolType string

const (
	BlockVolume VolType = "block"
)

type StorageVolumePhase string

const (
	VolumeCreated StorageVolumePhase = "Created"
	VolumeDeleted StorageVolumePhase = "Deleted"
)

type StorageVolumeSpec struct {
	// This field specifies whether this volume is block,
	// file or object store.
	VolumeType VolType `json:"volumetype"`

	// This field specifies the workload cluster ID.
	ClusterID string `json:"clusterid"`

	// This field specifies the storage pool used to
	// back this volume.
	PoolID string `json:"pool"`

	// This field specifies the filesystem of the volume
	// to be mounted
	// Defaults to ext4 if unspecified.
	FSType string `json:"fstype, omitempty"`

	// This field specifies the mount option type of the volume
	// to be mounted
	// Defaults to False if unspecified.
	ReadOnly bool `json:"readonly, omitempty"`

	// This field specifies whether data stored on this volume
	// should be deleted after the claim is removed.
	// Defaults to True if unspecified.
	Reclaim bool `json:"phase,omitempty"`
}

type StorageVolumeStatus struct {
	// Phase indicates state of volume creation or deletion
	Phase StorageVolumePhase

	// Message provides an explanation of the volume phase
	Message string `json:"message,omitempty"`

	// Reason provides an explanation of the last failure
	Reason string `json:"reason,omitempty"`
}

type StorageVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StorageVolumeSpec
	Status StorageVolumeStatus
}
