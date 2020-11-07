package v1

import (
	"fmt"

	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	rookclient "github.com/rook/rook/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

type BlockPoolInterface interface {
	Create(blockpool *StoragePool) error
	Update(blockpool *StoragePool) error
	Delete(pool string) error
	List() ([]StoragePool, error)
	Get(pool string) (*StoragePool, error)
}

type StoragePools struct {
	Namespace string
	Client    *rookclient.Clientset
}

func setupReplicatedSpec(pool *cephv1.CephBlockPool, blockpool *StoragePool) error {
	var replicationFactor uint
	var requireSafeReplicaSize bool

	requireSafeReplicaSize = true
	if blockpool.Spec.DurabilityPolicy.DurabilityLevel == DurabilityLevelLow {
		replicationFactor = 1
		requireSafeReplicaSize = false
	} else if blockpool.Spec.DurabilityPolicy.DurabilityLevel == DurabilityLevelSemi {
		replicationFactor = 2
	} else if blockpool.Spec.DurabilityPolicy.DurabilityLevel == DurabilityLevelNormal {
		replicationFactor = 3
	} else if blockpool.Spec.DurabilityPolicy.DurabilityLevel == DurabilityLevelHigh {
		replicationFactor = 4
	} else {
		ret := fmt.Errorf("No valid durability level specified, failed to create storage pool")
		return ret
	}
	pool.Spec.Replicated = cephv1.ReplicatedSpec{
		Size:                   replicationFactor,
		TargetSizeRatio:        1.0,
		RequireSafeReplicaSize: requireSafeReplicaSize,
	}
	return nil
}

func setupErasureCodedSpec(pool *cephv1.CephBlockPool, blockpool *StoragePool) error {
	var dataChunks uint
	var codingChunks uint

	if blockpool.Spec.DurabilityPolicy.DurabilityLevel == DurabilityLevelSemi {
		dataChunks = 2
		codingChunks = 1
	} else if blockpool.Spec.DurabilityPolicy.DurabilityLevel == DurabilityLevelNormal {
		dataChunks = 3
		codingChunks = 2
	} else if blockpool.Spec.DurabilityPolicy.DurabilityLevel == DurabilityLevelHigh {
		dataChunks = 4
		codingChunks = 3
	} else {
		ret := fmt.Errorf("No valid durability level specified, failed to create storage pool")
		return ret
	}
	pool.Spec.ErasureCoded = cephv1.ErasureCodedSpec{
		CodingChunks: codingChunks,
		DataChunks:   dataChunks,
	}

	return nil
}

func (p *StoragePools) Create(blockpool *StoragePool) error {
	var ret error
	var domain string
	var deviceClass string

	poolname := blockpool.ObjectMeta.Name
	clustername := blockpool.Spec.ClusterID
	rookclnt := p.Client
	_, err := rookclnt.CephV1().CephBlockPools(p.Namespace).Get(poolname, metav1.GetOptions{})
	if err == nil {
		ret = fmt.Errorf("Storage Pool already exists, cannot create blockpool")
		return ret
	} else {
		if blockpool.Spec.DurabilityPolicy.FailureDomain == FailureDomainHost {
			domain = "host"
		} else if blockpool.Spec.DurabilityPolicy.FailureDomain == FailureDomainRack {
			domain = "rack"
		} else {
			ret = fmt.Errorf("Invalid Failure domain specified, failed to create storage pool")
			return ret
		}
		if blockpool.Spec.PerfPolicy.IoPerfClass == DevStandard {
			deviceClass = "hdd"
		} else if blockpool.Spec.PerfPolicy.IoPerfClass == DevMedium {
			deviceClass = "ssd"
		} else if blockpool.Spec.PerfPolicy.IoPerfClass == DevFast {
			deviceClass = "nvme"
		}
		pool := &cephv1.CephBlockPool{
			ObjectMeta: metav1.ObjectMeta{
				Name:      poolname,
				Namespace: clustername,
			},
			Spec: cephv1.PoolSpec{
				FailureDomain:   domain,
				CrushRoot:       "",
				CompressionMode: "none",
				DeviceClass:     deviceClass,
			},
		}
		if blockpool.Spec.DurabilityPolicy.DurabilityClass == DurabilityClassReplicated {
			ret = setupReplicatedSpec(pool, blockpool)
			if ret != nil {
				return ret
			}
		} else if blockpool.Spec.DurabilityPolicy.DurabilityClass == DurabilityClassErasureCoded {
			ret = setupErasureCodedSpec(pool, blockpool)
			if ret != nil {
				return ret
			}
		} else {
			ret = fmt.Errorf("No valid durability class specified, failed to create storage pool")
			return ret
		}
		_, err = rookclnt.CephV1().CephBlockPools(clustername).Create(pool)
		if err == nil {
			fmt.Printf("Ceph Block pool created %s \n", poolname)
		} else {
			fmt.Printf("Failed to create Ceph block pool %v", err)
		}
	}

	return err
}

func (p *StoragePools) Update(blockpool *StoragePool) error {
	var domain string
	var ret error
	var deviceClass string

	rookclnt := p.Client
	poolname := blockpool.ObjectMeta.Name
	if blockpool.Spec.DurabilityPolicy.FailureDomain == FailureDomainHost {
		domain = "host"
	} else if blockpool.Spec.DurabilityPolicy.FailureDomain == FailureDomainRack {
		domain = "rack"
	} else {
		ret = fmt.Errorf("Invalid Failure domain specified, failed to create storage pool")
		return ret
	}
	if blockpool.Spec.PerfPolicy.IoPerfClass == DevStandard {
		deviceClass = "hdd"
	} else if blockpool.Spec.PerfPolicy.IoPerfClass == DevMedium {
		deviceClass = "ssd"
	} else if blockpool.Spec.PerfPolicy.IoPerfClass == DevFast {
		deviceClass = "nvme"
	}
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pool, err := rookclnt.CephV1().CephBlockPools(p.Namespace).Get(poolname, metav1.GetOptions{})
		if err == nil {
			pool.Spec.DeviceClass = deviceClass
			pool.Spec.FailureDomain = domain
			if blockpool.Spec.DurabilityPolicy.DurabilityClass == DurabilityClassReplicated && pool.Spec.Replicated.Size == 0 {
				ret = fmt.Errorf("Failed to update Ceph block pool, Invalid durability class specified")
				return ret
			}
			if blockpool.Spec.DurabilityPolicy.DurabilityClass == DurabilityClassErasureCoded && pool.Spec.Replicated.Size != 0 {
				ret = fmt.Errorf("Failed to update Ceph block pool, Invalid durability class specified")
				return ret
			}
			if blockpool.Spec.DurabilityPolicy.DurabilityClass == DurabilityClassReplicated {
				ret = setupReplicatedSpec(pool, blockpool)
				if ret != nil {
					return ret
				}
			} else if blockpool.Spec.DurabilityPolicy.DurabilityClass == DurabilityClassErasureCoded {
				ret = setupErasureCodedSpec(pool, blockpool)
				if ret != nil {
					return ret
				}
			} else {
				ret = fmt.Errorf("No valid durability class specified, failed to create storage pool")
				return ret
			}
			_, err = rookclnt.CephV1().CephBlockPools(p.Namespace).Update(pool)
			if err == nil {
				fmt.Printf("Ceph Block pool updated %s \n", poolname)
			}
		}
		return err
	})

	return err
}

func mapPoolPhase(pool *cephv1.CephBlockPool) StoragePoolPhase {
	if pool.Status.Phase == "Creating" {
		return PoolPhaseConnecting
	} else if pool.Status.Phase == "Ready" {
		return PoolPhaseReady
	} else if pool.Status.Phase == "Failure" {
		return PoolPhaseFailure
	} else if pool.Status.Phase == "Deleting" {
		return PoolPhaseDeleting
	}
	return ""
}

func (p *StoragePools) Get(poolname string) (*StoragePool, error) {
	var dPolicy StoragePolicyDurability
	var phase StoragePoolPhase
	var perfPolicy StoragePolicyPerformance

	rookclnt := p.Client
	pool, err := rookclnt.CephV1().CephBlockPools(p.Namespace).Get(poolname, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	phase = mapPoolPhase(pool)
	perfPolicy.ObjectMeta.Namespace = "storage-config"
	dPolicy.ObjectMeta.Namespace = "storage-config"
	if pool.Spec.DeviceClass == "nvme" {
		perfPolicy.ObjectMeta.Name = "sp-performance-fast"
		perfPolicy.IoPerfClass = DevFast
	} else if pool.Spec.DeviceClass == "ssd" {
		perfPolicy.ObjectMeta.Name = "sp-performance-medium"
		perfPolicy.IoPerfClass = DevMedium
	} else if pool.Spec.DeviceClass == "hdd" {
		perfPolicy.ObjectMeta.Name = "sp-performance-std"
		perfPolicy.IoPerfClass = DevStandard
	}
	if pool.Spec.FailureDomain == "host" {
		dPolicy.FailureDomain = FailureDomainHost
	} else if pool.Spec.FailureDomain == "rack" {
		dPolicy.FailureDomain = FailureDomainRack
	}
	if pool.Spec.Replicated.Size != 0 {
		dPolicy.DurabilityClass = "replicated"
		if pool.Spec.Replicated.Size == 1 {
			dPolicy.ObjectMeta.Name = "sp-durability-low"
			dPolicy.DurabilityLevel = DurabilityLevelLow
		} else if pool.Spec.Replicated.Size == 2 {
			dPolicy.ObjectMeta.Name = "sp-durability-semi"
			dPolicy.DurabilityLevel = DurabilityLevelSemi
		} else if pool.Spec.Replicated.Size == 3 {
			dPolicy.ObjectMeta.Name = "sp-durability-normal"
			dPolicy.DurabilityLevel = DurabilityLevelNormal
		} else if pool.Spec.Replicated.Size == 4 {
			dPolicy.ObjectMeta.Name = "sp-durability-high"
			dPolicy.DurabilityLevel = DurabilityLevelHigh
		}
	} else {
		dPolicy.DurabilityClass = "erasurecoded"
		if pool.Spec.ErasureCoded.CodingChunks == 1 {
			dPolicy.DurabilityLevel = DurabilityLevelSemi
		} else if pool.Spec.ErasureCoded.CodingChunks == 2 {
			dPolicy.DurabilityLevel = DurabilityLevelNormal
		} else if pool.Spec.ErasureCoded.CodingChunks == 3 {
			dPolicy.DurabilityLevel = DurabilityLevelHigh
		}
	}
	blockpool := &StoragePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      poolname,
			Namespace: p.Namespace,
		},
		Spec: StoragePoolSpec{
			ClusterID:        pool.ObjectMeta.Namespace,
			Quota:            0,
			DurabilityPolicy: dPolicy,
			PerfPolicy:       perfPolicy,
		},
		Status: StoragePoolStatus{
			Phase: phase,
		},
	}

	return blockpool, nil
}

func (p *StoragePools) Delete(poolname string) error {
	rookclnt := p.Client
	err := rookclnt.CephV1().CephBlockPools(p.Namespace).Delete(poolname, &metav1.DeleteOptions{})
	if err == nil {
		fmt.Printf("Ceph Block pool deleted %s \n", poolname)
	}
	return err
}

func (p *StoragePools) List() ([]StoragePool, error) {
	var plist []StoragePool

	return plist, nil
}
