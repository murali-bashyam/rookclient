package storageapi

import (
	rookclient "github.com/rook/rook/pkg/client/clientset/versioned"
	restclient "k8s.io/client-go/rest"
)

type Clientset struct {
	rookclnt *rookclient.Clientset
}

func NewForConfig(c *restclient.Config) (*Clientset, error) {
	var cs Clientset

	rookclnt, err := rookclient.NewForConfig(config)
	if err != nil {
		log.Println("Failed to initialize Rook Ceph clientset", err)
		return nil, err
	}
	cs.rookclnt = rookclnt
	return &cs, nil
}

func (c *Clientset) StorageClusters(ns string) *StorageClusters {
	return newStorageClusters(c.rookclnt, ns)
}

func (c *Clientset) StoragePools(ns string) *StoragePools {
	return newStoragePools(c.rookclnt, ns)
}

func (c *Clientset) StorageVolumes(ns string) *StorageVolumes {
	return newStorageVolumes(c.rookclnt, ns)
}
