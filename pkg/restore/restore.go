package restore

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func RestoreResources(backupDir, namespace string, clientset *kubernetes.Clientset) error {
	restoreFuncs := map[string]func(string, string, string, *kubernetes.Clientset) error{
		"pvc":            restorePVC,
		"pod":            restorePod,
		"replicaset":     restoreReplicaSet,
		"deployment":     restoreDeployment,
		"configmap":      restoreConfigMap,
		"service":        restoreServices,
		"statefulset":    restoreStatefulSet,
		"serviceaccount": restoreServiceAccounts,
		"secret":         restoreSecrets,
		// Add more resource types if needed
	}

	for resourceType, restoreFunc := range restoreFuncs {
		files, err := filepath.Glob(filepath.Join(backupDir, fmt.Sprintf("%s-*.json", resourceType)))
		if err != nil {
			return err
		}
		for _, file := range files {
			if err := restoreFunc(file, namespace, backupDir, clientset); err != nil {
				return err
			}
		}
	}

	return nil
}

func restorePVC(file, namespace, backupDir string, clientset *kubernetes.Clientset) error {
	ctx := context.Background()

	// List all PVCs in the namespace
	existingPVCs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Iterate through the PVC files in the backup directory
	pvcFiles, err := filepath.Glob(filepath.Join(backupDir, "pvc-*.json"))
	if err != nil {
		return err
	}

	for _, pvcFile := range pvcFiles {
		// Read the PVC JSON from the file
		pvcJSON, err := ioutil.ReadFile(pvcFile)
		if err != nil {
			return err
		}

		// Unmarshal the JSON into a PVC object
		var pvc corev1.PersistentVolumeClaim
		if err := json.Unmarshal(pvcJSON, &pvc); err != nil {
			return err
		}

		// Set the namespace of the restored PVC to match the requested namespace
		pvc.Namespace = namespace

		// Remove the resourceVersion field to avoid setting it when creating the PVC
		pvc.ResourceVersion = ""

		// Check if the PVC already exists in the namespace
		var exists bool
		for _, existingPVC := range existingPVCs.Items {
			if existingPVC.Name == pvc.Name {
				exists = true
				break
			}
		}

		// If the PVC already exists, skip restoring it
		if exists {
			continue
		}

		// Create the PVC
		_, err = clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, &pvc, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func restorePod(file, namespace, backupDir string, clientset *kubernetes.Clientset) error {
	ctx := context.Background()

	// List all Pods in the namespace
	existingPods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Iterate through the Pod files in the backup directory
	podFiles, err := filepath.Glob(filepath.Join(backupDir, "pod-*.json"))
	if err != nil {
		return err
	}

	for _, podFile := range podFiles {
		// Read the Pod JSON from the file
		podJSON, err := ioutil.ReadFile(podFile)
		if err != nil {
			return err
		}

		// Unmarshal the JSON into a Pod object
		var pod corev1.Pod
		if err := json.Unmarshal(podJSON, &pod); err != nil {
			return err
		}

		// Set the namespace of the restored Pod to match the requested namespace
		pod.Namespace = namespace
		// Remove the resourceVersion field to avoid setting it when creating the Pod
		pod.ResourceVersion = ""

		// Check if the Pod already exists in the namespace
		var exists bool
		for _, existingPod := range existingPods.Items {
			if existingPod.Name == pod.Name {
				exists = true
				break
			}
		}

		// If the Pod already exists, skip restoring it
		if exists {
			continue
		}

		// Create the Pod
		_, err = clientset.CoreV1().Pods(namespace).Create(ctx, &pod, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func restoreReplicaSet(file, namespace, backupDir string, clientset *kubernetes.Clientset) error {
	ctx := context.Background()

	// List all ReplicaSets in the namespace
	existingReplicaSets, err := clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Iterate through the ReplicaSet files in the backup directory
	rsFiles, err := filepath.Glob(filepath.Join(backupDir, "replicaset-*.json"))
	if err != nil {
		return err
	}

	for _, rsFile := range rsFiles {
		// Read the ReplicaSet JSON from the file
		rsJSON, err := ioutil.ReadFile(rsFile)
		if err != nil {
			return err
		}

		// Unmarshal the JSON into a ReplicaSet object
		var rs appsv1.ReplicaSet
		if err := json.Unmarshal(rsJSON, &rs); err != nil {
			return err
		}

		// Set the namespace of the restored ReplicaSet to match the requested namespace
		rs.Namespace = namespace

		// Remove the resourceVersion field to avoid setting it when creating the ReplicaSet
		rs.ResourceVersion = ""

		// Check if the ReplicaSet already exists in the namespace
		var exists bool
		for _, existingRS := range existingReplicaSets.Items {
			if existingRS.Name == rs.Name {
				exists = true
				break
			}
		}

		// If the ReplicaSet already exists, skip restoring it
		if exists {
			continue
		}

		// Create the ReplicaSet
		_, err = clientset.AppsV1().ReplicaSets(namespace).Create(ctx, &rs, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func restoreDeployment(file, namespace, backupDir string, clientset *kubernetes.Clientset) error {
	ctx := context.Background()

	// List all Deployments in the namespace
	existingDeployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Iterate through the Deployment files in the backup directory
	deploymentFiles, err := filepath.Glob(filepath.Join(backupDir, "deployment-*.json"))
	if err != nil {
		return err
	}

	for _, deploymentFile := range deploymentFiles {
		// Read the Deployment JSON from the file
		deploymentJSON, err := ioutil.ReadFile(deploymentFile)
		if err != nil {
			return err
		}

		// Unmarshal the JSON into a Deployment object
		var deployment appsv1.Deployment
		if err := json.Unmarshal(deploymentJSON, &deployment); err != nil {
			return err
		}

		// Set the namespace of the restored Deployment to match the requested namespace
		deployment.Namespace = namespace

		// Remove the resourceVersion field to avoid setting it when creating the Deployment
		deployment.ResourceVersion = ""

		// Check if the Deployment already exists in the namespace
		var exists bool
		for _, existingDeployment := range existingDeployments.Items {
			if existingDeployment.Name == deployment.Name {
				exists = true
				break
			}
		}

		// If the Deployment already exists, skip restoring it
		if exists {
			continue
		}

		// Create the Deployment
		_, err = clientset.AppsV1().Deployments(namespace).Create(ctx, &deployment, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func restoreConfigMap(file, namespace, backupDir string, clientset *kubernetes.Clientset) error {
	ctx := context.Background()

	// List all ConfigMaps in the namespace
	existingCMs, err := clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Iterate through the ConfigMap files in the backup directory
	cmFiles, err := filepath.Glob(filepath.Join(backupDir, "configmap-*.json"))
	if err != nil {
		return err
	}

	for _, cmFile := range cmFiles {
		// Read the ConfigMap JSON from the file
		cmJSON, err := ioutil.ReadFile(cmFile)
		if err != nil {
			return err
		}

		// Unmarshal the JSON into a ConfigMap object
		var cm corev1.ConfigMap
		if err := json.Unmarshal(cmJSON, &cm); err != nil {
			return err
		}

		// Check if the ConfigMap already exists in the namespace
		var exists bool
		for _, existingCM := range existingCMs.Items {
			if existingCM.Name == cm.Name {
				exists = true
				break
			}
		}

		// If the ConfigMap already exists, skip restoring it
		if exists {
			continue
		}

		// Create the ConfigMap
		_, err = clientset.CoreV1().ConfigMaps(namespace).Create(ctx, &cm, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func restoreStatefulSet(file, namespace, backupDir string, clientset *kubernetes.Clientset) error {
	ctx := context.Background()

	// List all StatefulSets in the namespace
	existingStatefulSets, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Iterate through the StatefulSet files in the backup directory
	statefulSetFiles, err := filepath.Glob(filepath.Join(backupDir, "statefulset-*.json"))
	if err != nil {
		return err
	}

	for _, statefulSetFile := range statefulSetFiles {
		// Read the StatefulSet JSON from the file
		statefulSetJSON, err := ioutil.ReadFile(statefulSetFile)
		if err != nil {
			return err
		}

		// Unmarshal the JSON into a StatefulSet object
		var statefulSet appsv1.StatefulSet
		if err := json.Unmarshal(statefulSetJSON, &statefulSet); err != nil {
			return err
		}

		// Set the namespace of the restored StatefulSet to match the requested namespace
		statefulSet.Namespace = namespace

		// Remove the resourceVersion field to avoid setting it when creating the StatefulSet
		statefulSet.ResourceVersion = ""

		// Check if the StatefulSet already exists in the namespace
		var exists bool
		for _, existingStatefulSet := range existingStatefulSets.Items {
			if existingStatefulSet.Name == statefulSet.Name {
				exists = true
				break
			}
		}

		// If the StatefulSet already exists, skip restoring it
		if exists {
			continue
		}

		// Create the StatefulSet
		_, err = clientset.AppsV1().StatefulSets(namespace).Create(ctx, &statefulSet, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func restoreServices(file, namespace, backupDir string, clientset *kubernetes.Clientset) error {
	ctx := context.Background()

	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "service-") {
			serviceJSON, err := ioutil.ReadFile(filepath.Join(backupDir, file.Name()))
			if err != nil {
				return err
			}

			var service corev1.Service
			if err := json.Unmarshal(serviceJSON, &service); err != nil {
				return err
			}

			// Set the namespace to the target namespace
			service.ObjectMeta.Namespace = namespace

			// Remove resourceVersion field
			service.ObjectMeta.ResourceVersion = ""

			// Unset the IP to allow dynamic allocation
			service.Spec.ClusterIP = ""

			// Remove the clusterIPs field
			service.Spec.ClusterIPs = nil

			// Check if the service already exists
			_, err = clientset.CoreV1().Services(namespace).Get(ctx, service.Name, metav1.GetOptions{})
			if err == nil {
				// Service already exists, skip creation
				continue
			} else if !errors.IsNotFound(err) {
				// Unexpected error occurred
				return err
			}

			// Service does not exist, create it
			_, err = clientset.CoreV1().Services(namespace).Create(ctx, &service, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func restoreServiceAccounts(file, namespace, backupDir string, clientset *kubernetes.Clientset) error {
	ctx := context.Background()

	// Iterate through backup files
	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return err
	}

	// Restore each ServiceAccount from backup files
	for _, file := range files {
		// Read backup file
		data, err := ioutil.ReadFile(filepath.Join(backupDir, file.Name()))
		if err != nil {
			return err
		}

		// Unmarshal JSON data into ServiceAccount object
		var sa corev1.ServiceAccount
		if err := json.Unmarshal(data, &sa); err != nil {
			return err
		}

		// Check if the ServiceAccount already exists
		_, err = clientset.CoreV1().ServiceAccounts(namespace).Get(ctx, sa.Name, metav1.GetOptions{})
		if err == nil {
			// ServiceAccount already exists, skip
			continue
		} else if !errors.IsNotFound(err) {
			// An error occurred other than "not found"
			return err
		}

		// Set the namespace to the target namespace
		sa.Namespace = namespace
		sa.ObjectMeta.ResourceVersion = ""

		// Create the ServiceAccount
		_, err = clientset.CoreV1().ServiceAccounts(namespace).Create(ctx, &sa, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func restoreSecrets(file, namespace, backupDir string, clientset *kubernetes.Clientset) error {
	ctx := context.Background()

	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "secret-") {
			secretJSON, err := ioutil.ReadFile(filepath.Join(backupDir, file.Name()))
			if err != nil {
				return err
			}

			var secret corev1.Secret
			if err := json.Unmarshal(secretJSON, &secret); err != nil {
				return err
			}

			// Set the namespace to the target namespace
			secret.ObjectMeta.Namespace = namespace

			// Remove resourceVersion field
			secret.ObjectMeta.ResourceVersion = ""

			// Check if the secret already exists
			_, err = clientset.CoreV1().Secrets(namespace).Get(ctx, secret.Name, metav1.GetOptions{})
			if err == nil {
				// Secret already exists, skip creation
				continue
			} else if !errors.IsNotFound(err) {
				// Unexpected error occurred
				return err
			}

			// Secret does not exist, create it
			_, err = clientset.CoreV1().Secrets(namespace).Create(ctx, &secret, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
