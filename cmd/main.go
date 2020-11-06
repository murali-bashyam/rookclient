package main

import (
	"flag"
	"fmt"
	"log"

	storageapi "github.com/murali-bashyam/rookclient/pkg/storageapi"
	storageapiv1 "github.com/murali-bashyam/rookclient/pkg/storageapi/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func createCluster(c *storageapi.Clientset, ns string, clustername string) error {
	cluster := &storageapiv1.StorageCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername,
			Namespace: ns,
		},
	}
	cluster, err := c.StorageClusters(ns).Create(cluster)
	if err == nil {
		fmt.Printf("Ceph cluster created, status %s, phase %s message %s ",
			cluster.Status.State,
			cluster.Status.Phase,
			cluster.Status.Message)
	} else {
		fmt.Printf("Failed to create Ceph cluster, status %s, phase %s error %s ",
			cluster.Status.State,
			cluster.Status.Phase,
			cluster.Status.Message)
	}

	return err
}

func isClusterHealthy(c *storageapi.Clientset, ns string, clustername string) error {
	cluster, err := c.StorageClusters(ns).Get(clustername)
	if err == nil {
		if cluster.Status.State == storageapiv1.ClusterStateCreated &&
			cluster.Status.Phase == storageapiv1.ClusterPhaseReady {
			fmt.Printf("Cluster %s is healthy", ns)
		}
		fmt.Printf("Ceph cluster health status %s, phase %s message %s ",
			cluster.Status.State,
			cluster.Status.Phase,
			cluster.Status.Message)
	} else {
		fmt.Printf("Failed to get Ceph cluster info, status %s, phase %s error %s ",
			cluster.Status.State,
			cluster.Status.Phase,
			cluster.Status.Message)
	}

	return err
}

func createStoragePool(c *storageapi.Clientset, ns string, clustername string,
	poolname string) error {
	dpolicy := storageapiv1.StoragePolicyDurability{
		FailureDomain:   storageapiv1.FailureDomainHost,
		DurabilityClass: storageapiv1.DurabilityClassReplicated,
		DurabilityLevel: storageapiv1.DurabilityLevelNormal,
	}

	pool := &storageapiv1.StoragePool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      poolname,
			Namespace: ns,
		},
		Spec: storageapiv1.StoragePoolSpec{
			ClusterID:        clustername,
			Quota:            0,
			DurabilityPolicy: dpolicy,
		},
	}
	err := c.StoragePools(ns).Create(pool)
	if err != nil {
		fmt.Printf("Failed to create storage pool %v", err)
		return err
	}
	pool, err = c.StoragePools(ns).Get(poolname)
	if err != nil {
		fmt.Printf("Failed to get storage pool info %v", err)
	}
	return err
}

func deleteStoragePool(c *storageapi.Clientset, ns string, poolname string) error {
	err := c.StoragePools(ns).Delete(poolname)
	if err != nil {
		fmt.Printf("Failed to delete storage pool %v", err)
	}
	return err
}

func deleteStorageCluster(c *storageapi.Clientset, ns string, clustername string) error {
	err := c.StorageClusters(ns).Delete(clustername)
	if err != nil {
		fmt.Printf("Failed to delete storage cluster %v", err)
	}
	return err
}

func createBlockStorageVolume(c *storageapi.Clientset, ns string,
	clustername string, poolname string) error {
	volume := &storageapiv1.StorageVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      poolname,
			Namespace: ns,
		},
		Spec: storageapiv1.StorageVolumeSpec{
			VolumeType: storageapiv1.BlockVolume,
			PoolID:     poolname,
			ClusterID:  ns,
			FSType:     "ext4",
			ReadOnly:   false,
			Reclaim:    true,
		},
	}
	volume, sc, err := c.StorageVolumes(ns).Create(volume)
	if err != nil {
		fmt.Printf("Failed to create storage volume %v", err)
	} else {
		fmt.Println("Storage class for block volume : ")
		fmt.Printf("%s", *sc)
	}
	return err
}

func main() {

	var kubeconfig *string
	var config *restclient.Config

	kubeconfig = flag.String("kubeconfig", "/home/mbcoder/kubeconfig", "kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal("Failed to build config", err)
	}

	storageclnt, err := storageapi.NewForConfig(config)
	if err != nil {
		log.Fatal("Failed to build storage clientconfig", err)
	}
	err = createCluster(storageclnt, "rook-ceph", "rook-ceph")
	if err != nil {
		log.Fatal("Failed to create Ceph cluster ", err)
	} else {
		fmt.Printf("Successfully created Ceph cluster")
	}
	err = createStoragePool(storageclnt, "rook-ceph", "rook-ceph", "bpool")
	if err != nil {
		log.Fatal("Failed to create storage pool ", err)
	} else {
		fmt.Printf("Successfully created storage pool")
	}
	err = createBlockStorageVolume(storageclnt, "rook-ceph", "rook-ceph", "bpool")
	if err != nil {
		log.Fatal("Failed to create block storage volume ", err)
	} else {
		fmt.Printf("Successfully created block storage volume ")
	}
}
