package v1

import (
	"fmt"

	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	rookv1 "github.com/rook/rook/pkg/apis/rook.io/v1"
	rookclient "github.com/rook/rook/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterInterface interface {
	Create(cluster *StorageCluster) (*StorageCluster, error)
	Update(cluster *StorageCluster) (*StorageCluster, error)
	Delete(clustername string) error
	List() ([]StorageCluster, error)
	Get(clustername string) (*StorageCluster, error)
}

const (
	cephversion string = "ceph/v15.2.4"
	dirhostpath string = "/var/lib/rook"
)

type StorageClusters struct {
	Namespace string
	Client    *rookclient.Clientset
}

func (c *StorageClusters) Create(cluster *StorageCluster) (*StorageCluster, error) {
	var external bool

	rookclnt := c.Client
	useAllDevices := true
	cephcluster := &cephv1.CephCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: cluster.ObjectMeta.Name,
			Namespace: cluster.ObjectMeta.Namespace,
		},
		Spec: cephv1.ClusterSpec{
			DataDirHostPath: dirhostpath,
			Mon: cephv1.MonSpec{
				Count:                3,
				AllowMultiplePerNode: false,
			},
			Storage: rookv1.StorageScopeSpec{
				UseAllNodes: true,
				Selection: rookv1.Selection{
					UseAllDevices: &useAllDevices,
				},
			},
			Monitoring: cephv1.MonitoringSpec{
				Enabled: true,
			},
		},
	}
	if len(cluster.Spec.StorageClusterID) != 0 {
		cephcluster.Spec.External.Enable = external
	} else {
		cephcluster.Spec.CephVersion.Image = cephversion
	}
	cephcluster, err := rookclnt.CephV1().CephClusters(c.Namespace).Create(cephcluster)
	if err == nil {
		fmt.Printf("Ceph Cluster created %s \n", cluster.ObjectMeta.Name)
	}
	cluster.Status.Phase = mapClusterPhase(cephcluster)
	cluster.Status.State = mapClusterState(cephcluster)
	cluster.Status.Message = cephcluster.Status.Message
	return cluster, err
}

func (c *StorageClusters) Update(cluster *StorageCluster) (*StorageCluster, error) {
	return nil, nil
}

func (c *StorageClusters) Delete(clustername string) error {
	rookclnt := c.Client
	err := rookclnt.CephV1().CephClusters(c.Namespace).Delete(clustername, &metav1.DeleteOptions{})
	if err == nil {
		fmt.Printf("Ceph cluster deleted %s \n", clustername)
	}
	return err
}

func mapClusterPhase(c *cephv1.CephCluster) StorageClusterPhase {
	if c.Status.Phase == cephv1.ConditionIgnored {
		return ClusterPhaseIgnored
	} else if c.Status.Phase == cephv1.ConditionConnecting {
		return ClusterPhaseConnecting
	} else if c.Status.Phase == cephv1.ConditionConnected {
		return ClusterPhaseConnected
	} else if c.Status.Phase == cephv1.ConditionProgressing {
		return ClusterPhaseProgressing
	} else if c.Status.Phase == cephv1.ConditionReady {
		return ClusterPhaseReady
	} else if c.Status.Phase == cephv1.ConditionUpdating {
		return ClusterPhaseUpdating
	} else if c.Status.Phase == cephv1.ConditionFailure {
		return ClusterPhaseFailure
	} else if c.Status.Phase == cephv1.ConditionUpgrading {
		return ClusterPhaseUpgrading
	} else if c.Status.Phase == cephv1.ConditionDeleting {
		return ClusterPhaseDeleting
	}
	return ""
}

func mapClusterState(c *cephv1.CephCluster) StorageClusterState {
	if c.Status.State == cephv1.ClusterStateCreating {
		return ClusterStateCreating
	} else if c.Status.State == cephv1.ClusterStateCreated {
		return ClusterStateCreated
	} else if c.Status.State == cephv1.ClusterStateUpdating {
		return ClusterStateUpdating
	} else if c.Status.State == cephv1.ClusterStateConnecting {
		return ClusterStateConnecting
	} else if c.Status.State == cephv1.ClusterStateConnected {
		return ClusterStateConnected
	} else if c.Status.State == cephv1.ClusterStateError {
		return ClusterStateError
	}
	return ""
}

func (c *StorageClusters) Get(clustername string) (*StorageCluster, error) {
	var monitoring bool
	var externalClusterID string
	var phase StorageClusterPhase
	var state StorageClusterState

	rookclnt := c.Client
	cephcluster, err := rookclnt.CephV1().CephClusters(c.Namespace).Get(clustername, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if cephcluster.Spec.Monitoring.Enabled == true {
		monitoring = true
	}

	if cephcluster.Spec.External.Enable == true {
		externalClusterID = "external-storage-cluster"
	}

	phase = mapClusterPhase(cephcluster)
	state = mapClusterState(cephcluster)
	cluster := &StorageCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: cephcluster.ObjectMeta.Name,
		},
		Spec: StorageClusterSpec{
			StorageClusterID: externalClusterID,
			Monitoring:       monitoring,
		},
		Status: StorageClusterStatus{
			Phase:   phase,
			State:   state,
			Message: cephcluster.Status.Message,
		},
	}

	return cluster, nil
}

func (c *StorageClusters) List() ([]StorageCluster, error) {
	var clist []StorageCluster

	return clist, nil
}
