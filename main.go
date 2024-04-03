package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	// Add this import
	appsv1 "k8s.io/api/apps/v1" // Add this import
	corev1 "k8s.io/api/core/v1" // Add this import
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Application struct {
	AppID     string `json:"app_id"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type Backup struct {
	BackupID string `json:"backup_id"`
	AppID    string `json:"app_id"`
}

var appCounter int = 0
var backupCounter int = 0
var apps map[string]Application = make(map[string]Application)
var appNameNamespaceMap map[string]string = make(map[string]string)
var backups map[string]Backup = make(map[string]Backup)

var clientset *kubernetes.Clientset // Declare clientset as a global variable

func main() {
	// Set the KUBECONFIG environment variable to point to the kubeconfig file
	kubeconfig := os.Getenv("HOME") + "/.kube/config"
	os.Setenv("KUBECONFIG", kubeconfig)

	// Initialize Kubernetes clientset using kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	router := gin.Default()

	router.PUT("/application/", defineApplication)
	router.POST("/backup", performBackup)
	router.PUT("/restore", restoreBackup)

	router.Run(":8080")
}

func defineApplication(c *gin.Context) {
	var app Application
	if err := c.BindJSON(&app); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the combination of app name and namespace already exists
	appNameNamespaceKey := fmt.Sprintf("%s_%s", app.Name, app.Namespace)
	if existingAppID, ok := appNameNamespaceMap[appNameNamespaceKey]; ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Application with same name and namespace already exists", "existing_app_id": existingAppID})
		return
	}

	// Increment appCounter for app_id
	appCounter++
	appID := fmt.Sprintf("app_%d", appCounter)

	// Store the application in both maps
	app.AppID = appID // Include the app_id in the Application struct

	apps[appID] = app
	appNameNamespaceMap[appNameNamespaceKey] = appID

	c.JSON(http.StatusOK, gin.H{"app_id": appID})
}

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

func performBackup(c *gin.Context) {
	var requestBody struct {
		AppID string `json:"app_id"`
	}

	// Parse JSON request body
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve the application details using the provided app ID
	app, ok := apps[requestBody.AppID]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid app_id"})
		return
	}

	// Generate a unique backup ID
	backupCounter++
	backupID := fmt.Sprintf("backup_%d", backupCounter)

	// Perform backup operations for relevant resources (PVC, Pod, ReplicaSet, Deployment, ConfigMap, etc.)

	// Create a directory to store the backup files
	backupDir := fmt.Sprintf("./backups/%s", backupID)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Perform backup operations for relevant resources
	if err := BackupPVCs(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := BackupPods(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := BackupReplicaSets(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := BackupDeployments(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := BackupConfigMaps(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := BackupStatefulSet(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := BackupServices(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := BackupServiceAccounts(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := BackupSecrets(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Associate the backup ID with the app ID for future reference
	backup := Backup{
		BackupID: backupID,
		AppID:    app.AppID,
	}
	backups[backupID] = backup

	// Return response
	c.JSON(http.StatusOK, gin.H{"backup_id": backupID, "app_id": app.AppID})
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

func restoreBackup(c *gin.Context) {
	var requestBody struct {
		Namespace string `json:"namespace"`
		BackupID  string `json:"backup_id"`
	}

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the context from gin.Context
	ctx := c.Request.Context()

	// Validate if the namespace exists
	_, err := clientset.CoreV1().Namespaces().Get(ctx, requestBody.Namespace, metav1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace does not exist"})
		return
	}

	// Get the backup directory
	backupDir := fmt.Sprintf("./backups/%s", requestBody.BackupID)

	// Restore resources
	if err := restoreResources(backupDir, requestBody.Namespace); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Restore completed successfully"})
}

func restoreResources(backupDir, namespace string) error {
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
