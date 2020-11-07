package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	storageapi "github.com/murali-bashyam/rookclient/pkg/storageapi"
	storageapiv1 "github.com/murali-bashyam/rookclient/pkg/storageapi/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	ROOK_NAMESPACE    string = "rook-ceph"
	CLUSTER_NAME      string = "rook-ceph"
	STORAGE_POOL_NAME string = "bpool1"
	KUBE_CONFIG       string = "/home/mbcoder/kubeconfig"
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
		fmt.Printf("Storage cluster created, status %s, phase %s message %s \n",
			cluster.Status.State,
			cluster.Status.Phase,
			cluster.Status.Message)
	} else {
		fmt.Printf("Failed to create Storage cluster, status %s, phase %s error %s \n",
			cluster.Status.State,
			cluster.Status.Phase,
			cluster.Status.Message)
	}

	return err
}

func isClusterHealthy(c *storageapi.Clientset, ns string, clustername string) (bool, error) {
	cluster, err := c.StorageClusters(ns).Get(clustername)
	if err == nil {
		if cluster.Status.State == storageapiv1.ClusterStateCreated &&
			cluster.Status.Phase == storageapiv1.ClusterPhaseReady {
			fmt.Printf("Storage Cluster %s is healthy \n", ns)
			return true, nil
		}
		fmt.Printf("Storage cluster health status %s, phase %s message %s \n",
			cluster.Status.State,
			cluster.Status.Phase,
			cluster.Status.Message)
	} else {
		fmt.Printf("Failed to get storage cluster info, %s \n", clustername)
	}
	return false, err
}

func isStoragePoolReady(c *storageapi.Clientset, ns string, clustername string, poolname string) (bool, error) {
	pool, err := c.StoragePools(ns).Get(poolname)
	if err == nil {
		if pool != nil {
			if pool.Status.Phase == storageapiv1.PoolPhaseReady {
				fmt.Printf("Storage Pool %s is ready \n", poolname)
				return true, nil
			}
			fmt.Printf("Storage pool status %s \n", pool.Status.Phase)
		}
	} else {
		fmt.Printf("Failed to get storage pool info, %s \n", poolname)
	}
	return false, err
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
		fmt.Printf("Failed to create storage pool %v \n", err)
		return err
	}
	pool, err = c.StoragePools(ns).Get(poolname)
	if err != nil {
		fmt.Printf("Failed to get storage pool info %v \n", err)
	}
	return err
}

func deleteStoragePool(c *storageapi.Clientset, ns string, poolname string) error {
	err := c.StoragePools(ns).Delete(poolname)
	if err != nil {
		fmt.Printf("Failed to delete storage pool %v \n", err)
	}
	return err
}

func deleteStorageCluster(c *storageapi.Clientset, ns string, clustername string) error {
	err := c.StorageClusters(ns).Delete(clustername)
	if err != nil {
		fmt.Printf("Failed to delete storage cluster %v\n", err)
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
			ClusterID:  clustername,
			FSType:     "ext4",
			ReadOnly:   false,
			Reclaim:    true,
		},
	}
	volume, sc, err := c.StorageVolumes(ns).Create(volume)
	if err != nil {
		fmt.Printf("Failed to create storage volume %v \n", err)
	} else {
		fmt.Println("Storage class for block volume : ")
		fmt.Printf("%s \n", *sc)
	}
	return err
}

func main() {

	var kubeconfig *string
	var config *restclient.Config
	var incluster bool

	if len(os.Args) > 1 {
		if os.Args[1] == "incluster" {
			incluster = true
		}
	}

	if incluster == false {
		kubeconfig = flag.String("kubeconfig", KUBE_CONFIG, "kubeconfig file")
		flag.Parse()

		kconfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			log.Fatal("Failed to build config", err)
		}
		config = kconfig
	} else {
		kconfig, err := restclient.InClusterConfig()
		if err != nil {
			log.Fatal("Failed to build config", err)
		} else {
			fmt.Println("Successfully built incluster config")
		}
		config = kconfig
	}

	storageclnt, err := storageapi.NewForConfig(config)
	if err != nil {
		log.Fatal("Failed to build storage clientconfig", err)
	}
	err = createCluster(storageclnt, ROOK_NAMESPACE, CLUSTER_NAME)
	if err != nil {
		log.Fatal("Failed to create storage cluster ", err)
	} else {
		fmt.Printf("Successfully created storage cluster %s \n", CLUSTER_NAME)
	}
	for {
		status, err := isClusterHealthy(storageclnt, ROOK_NAMESPACE, CLUSTER_NAME)
		if err != nil {
			log.Fatal("Failed to determine health of storage cluster \n", err)
		} else if status == true {
			fmt.Printf("Storage cluster is healthy %s \n", CLUSTER_NAME)
			break
		} else {
			fmt.Printf("Waiting for Storage cluster to become healthy %s \n", CLUSTER_NAME)
		}
		time.Sleep(10 * time.Second)
	}
	err = createStoragePool(storageclnt, ROOK_NAMESPACE, CLUSTER_NAME, STORAGE_POOL_NAME)
	if err != nil {
		log.Fatal("Failed to create storage pool \n", err)
	} else {
		fmt.Printf("Successfully created storage pool %s \n", STORAGE_POOL_NAME)
	}

	for {
		status, err := isStoragePoolReady(storageclnt, ROOK_NAMESPACE, CLUSTER_NAME, STORAGE_POOL_NAME)
		if err != nil {
			log.Fatal("Failed to determine status of storage pool \n", err)
		} else if status == true {
			fmt.Printf("Storage pool is ready %s \n", STORAGE_POOL_NAME)
			break
		} else {
			fmt.Printf("Waiting for Storage pool to become ready %s \n", STORAGE_POOL_NAME)
		}
		time.Sleep(10 * time.Second)
	}
	err = createBlockStorageVolume(storageclnt, ROOK_NAMESPACE, CLUSTER_NAME, STORAGE_POOL_NAME)
	if err != nil {
		log.Fatal("Failed to create block storage volume \n", err)
	} else {
		fmt.Printf("Successfully created block storage volume \n")
	}

	if incluster == true {
		for {
			status, err := isClusterHealthy(storageclnt, ROOK_NAMESPACE, CLUSTER_NAME)
			if err != nil {
				log.Fatal("Failed to determine health of storage cluster ", err)
			} else if status == true {
				fmt.Printf("Storage cluster is healthy %s \n", CLUSTER_NAME)
			} else {
				fmt.Printf("Waiting for Storage cluster to become healthy %s \n", CLUSTER_NAME)
			}
			time.Sleep(10 * time.Second)
		}
	}
}
