package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func BackupPVCs(clientset *kubernetes.Clientset, namespace, backupDir string) error {
	// Retrieve PVCs in the namespace
	pvcList, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Backup each PVC
	for _, pvc := range pvcList.Items {
		// Marshal PVC object to JSON
		pvcJSON, err := json.MarshalIndent(pvc, "", "  ")
		if err != nil {
			return err
		}

		// Write PVC JSON to file
		filename := filepath.Join(backupDir, fmt.Sprintf("%s.json", pvc.Name))
		if err := os.WriteFile(filename, pvcJSON, 0644); err != nil {
			return err
		}
	}

	return nil
}

func BackupPods(clientset *kubernetes.Clientset, namespace, backupDir string) error {
	podList, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pod := range podList.Items {
		podJSON, err := json.MarshalIndent(pod, "", "  ")
		if err != nil {
			return err
		}
		filename := filepath.Join(backupDir, fmt.Sprintf("pod-%s.json", pod.Name))
		if err := os.WriteFile(filename, podJSON, 0644); err != nil {
			return err
		}
	}
	return nil
}

func BackupSecrets(clientset *kubernetes.Clientset, namespace, backupDir string) error {
	ctx := context.Background()

	secretsList, err := clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, secret := range secretsList.Items {
		// Marshal Secret object to JSON
		secretJSON, err := json.MarshalIndent(secret, "", "  ")
		if err != nil {
			return err
		}

		// Write Secret JSON to file
		filename := filepath.Join(backupDir, "secret-"+secret.Name+".json")
		if err := os.WriteFile(filename, secretJSON, 0644); err != nil {
			return err
		}
	}
	return nil
}

func BackupReplicaSets(clientset *kubernetes.Clientset, namespace, backupDir string) error {
	rsList, err := clientset.AppsV1().ReplicaSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, rs := range rsList.Items {
		rsJSON, err := json.MarshalIndent(rs, "", "  ")
		if err != nil {
			return err
		}
		filename := filepath.Join(backupDir, fmt.Sprintf("replicaset-%s.json", rs.Name))
		if err := os.WriteFile(filename, rsJSON, 0644); err != nil {
			return err
		}
	}
	return nil
}

func BackupDeployments(clientset *kubernetes.Clientset, namespace, backupDir string) error {
	deploymentList, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, deployment := range deploymentList.Items {
		deploymentJSON, err := json.MarshalIndent(deployment, "", "  ")
		if err != nil {
			return err
		}
		filename := filepath.Join(backupDir, fmt.Sprintf("deployment-%s.json", deployment.Name))
		if err := os.WriteFile(filename, deploymentJSON, 0644); err != nil {
			return err
		}
	}
	return nil
}

func BackupConfigMaps(clientset *kubernetes.Clientset, namespace, backupDir string) error {
	ctx := context.Background()

	cmList, err := clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, cm := range cmList.Items {
		if cm.Name == "kube-root-ca.crt" {
			continue
		}

		// Check if ConfigMap already exists in backup directory
		filename := filepath.Join(backupDir, fmt.Sprintf("configmap-%s.json", cm.Name))
		if _, err := os.Stat(filename); err == nil {
			// Skip if ConfigMap already exists in backup directory
			continue
		}

		// Omit namespace and resourceVersion fields
		cm.ObjectMeta.Namespace = ""
		cm.ObjectMeta.ResourceVersion = ""

		cmJSON, err := json.MarshalIndent(cm, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filename, cmJSON, 0644); err != nil {
			return err
		}
	}
	return nil
}

func BackupStatefulSet(clientset *kubernetes.Clientset, namespace, backupDir string) error {
	ctx := context.Background()

	statefulSetList, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, statefulSet := range statefulSetList.Items {
		// Check if StatefulSet already exists in backup directory
		filename := filepath.Join(backupDir, fmt.Sprintf("statefulset-%s.json", statefulSet.Name))
		if _, err := os.Stat(filename); err == nil {
			// Skip if StatefulSet already exists in backup directory
			continue
		}

		// Omit namespace and resourceVersion fields
		statefulSet.ObjectMeta.Namespace = ""
		statefulSet.ObjectMeta.ResourceVersion = ""

		statefulSetJSON, err := json.MarshalIndent(statefulSet, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filename, statefulSetJSON, 0644); err != nil {
			return err
		}
	}
	return nil
}

func BackupServices(clientset *kubernetes.Clientset, namespace, backupDir string) error {
	ctx := context.Background()

	serviceList, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, service := range serviceList.Items {
		// Check if Service already exists in backup directory
		filename := filepath.Join(backupDir, fmt.Sprintf("service-%s.json", service.Name))
		if _, err := os.Stat(filename); err == nil {
			// Skip if Service already exists in backup directory
			continue
		}

		// Omit namespace and resourceVersion fields
		service.ObjectMeta.Namespace = ""
		service.ObjectMeta.ResourceVersion = ""

		serviceJSON, err := json.MarshalIndent(service, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filename, serviceJSON, 0644); err != nil {
			return err
		}
	}
	return nil
}

func BackupServiceAccounts(clientset *kubernetes.Clientset, namespace, backupDir string) error {
	ctx := context.Background()

	// Retrieve ServiceAccounts in the namespace
	saList, err := clientset.CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Backup each ServiceAccount
	for _, sa := range saList.Items {
		// Marshal ServiceAccount object to JSON
		saJSON, err := json.MarshalIndent(sa, "", "  ")
		if err != nil {
			return err
		}

		// Write ServiceAccount JSON to file
		filename := filepath.Join(backupDir, fmt.Sprintf("serviceaccount-%s.json", sa.Name))
		if err := os.WriteFile(filename, saJSON, 0644); err != nil {
			return err
		}
	}
	return nil
}
