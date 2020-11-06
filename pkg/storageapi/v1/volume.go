package v1

import (
	"fmt"

	rookclient "github.com/rook/rook/pkg/client/clientset/versioned"
)

type VolumeInterface interface {
	Create(volume *StorageVolume) (*StorageVolume, *string, error)
	Delete(volumename string) error
}

type StorageVolumes struct {
	Namespace string
	Client    *rookclient.Clientset
}

func createBlockStorageClass(volume *StorageVolume) string {
	var reclaimPolicy string

	storageClassName := volume.ObjectMeta.Name + "-block"
	poolName := volume.Spec.PoolID
	if volume.Spec.Reclaim == true {
		reclaimPolicy = "Delete"
	} else {
		reclaimPolicy = "Retain"
	}
	clusterid := volume.Spec.ClusterID
	namespace := volume.ObjectMeta.Namespace

	return `
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ` + storageClassName + `
provisioner: ` + namespace + `.rbd.csi.ceph.com
reclaimPolicy: ` + reclaimPolicy + `
parameters:
  clusterID: ` + clusterid + `
  pool: ` + poolName + `
  imageFormat: 2
  imageFeatures: layering
  csi.storage.k8s.io/provisioner-secret-name: rook-csi-rbd-provisioner
  csi.storage.k8s.io/provisioner-secret-namespace: ` + namespace + `
  csi.storage.k8s.io/controller-expand-secret-name: rook-csi-rbd-provisioner
  csi.storage.k8s.io/controller-expand-secret-namespace: ` + namespace + `
  csi.storage.k8s.io/node-stage-secret-name: rook-csi-rbd-node
  csi.storage.k8s.io/node-stage-secret-namespace: ` + namespace + `
  csi.storage.k8s.io/fstype: ext4
allowVolumeExpansion: true
`
}

func (s *StorageVolumes) Create(volume *StorageVolume) (*StorageVolume, *string, error) {
	var sc string

	if volume.Spec.VolumeType != BlockVolume {
		err := fmt.Errorf(" Invalid volume type, cannot create volume")
		return nil, &sc, err
	}

	sc = createBlockStorageClass(volume)
	volume.Status.Phase = VolumeCreated
	return volume, &sc, nil
}

func (s *StorageVolumes) Update(volume *StorageVolume) (*StorageVolume, error) {
	return nil, nil
}

func (s *StorageVolumes) Delete(volumename string) error {
	return nil
}

func (s *StorageVolumes) Get(volumename string) (*StorageVolume, error) {
	return nil, nil
}

func (s *StorageVolumes) List() ([]StorageVolume, error) {
	var vlist []StorageVolume

	return vlist, nil
}
