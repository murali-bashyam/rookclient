package storageapi

import (
	"fmt"

	storageapiv1 "github.com/murali-bashyam/rookclient/pkg/storageapi/v1"
	rookclient "github.com/rook/rook/pkg/client/clientset/versioned"
	restclient "k8s.io/client-go/rest"
)

type Clientset struct {
	rookclnt *rookclient.Clientset
}

func NewForConfig(config *restclient.Config) (*Clientset, error) {
	var cs Clientset

	rookclnt, err := rookclient.NewForConfig(config)
	if err != nil {
		fmt.Printf("Failed to initialize Rook Ceph clientset %v", err)
		return nil, err
	}
	cs.rookclnt = rookclnt
	return &cs, nil
}

func (c *Clientset) StorageClusters(namespace string) *storageapiv1.StorageClusters {
	return &storageapiv1.StorageClusters{
		Namespace: namespace,
		Client:    c.rookclnt,
	}
}

func (c *Clientset) StoragePools(namespace string) *storageapiv1.StoragePools {
	return &storageapiv1.StoragePools{
		Namespace: namespace,
		Client:    c.rookclnt,
	}
}

func (c *Clientset) StorageVolumes(namespace string) *storageapiv1.StorageVolumes {
	return &storageapiv1.StorageVolumes{
		Namespace: namespace,
		Client:    c.rookclnt,
	}
}
