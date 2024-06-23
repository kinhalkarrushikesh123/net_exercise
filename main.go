package main

import (
	"fmt"
	"net/http"
	"os"

	"net_exercise/pkg/backup"
	"net_exercise/pkg/restore"

	"github.com/gin-gonic/gin"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

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

	router.PUT("/application", defineApplication)
	router.PUT("/backup", performBackup)
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

	// Create a directory to store the backup files
	backupDir := fmt.Sprintf("./backups/%s", backupID)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Perform backup operations for relevant resources
	if err := backup.BackupPVCs(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := backup.BackupPods(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := backup.BackupReplicaSets(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := backup.BackupDeployments(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := backup.BackupConfigMaps(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := backup.BackupStatefulSet(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := backup.BackupServices(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := backup.BackupServiceAccounts(clientset, app.Namespace, backupDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := backup.BackupSecrets(clientset, app.Namespace, backupDir); err != nil {
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
	if err := restore.RestoreResources(backupDir, requestBody.Namespace, clientset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Restore completed successfully"})
}
