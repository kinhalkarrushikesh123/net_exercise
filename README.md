# NET_EXERCISE
exercise net k8s app namespace backup

This App Backs up all k8s resource of a Namespace and then Deploys that resource in a new Namespace


## Installing MariaDB with Helm

You can install MariaDB using Helm by following these steps:

1. Add the Bitnami Helm repository:

   ```bash
   helm repo add bitnami https://charts.bitnami.com/bitnami

2. Install app
   ```bash
   helm install --create-namespace -n test-mariadb mariadb --set auth.rootPassword=secret,auth.database=db bitnami/mariadb

This will install App in test-mariadb namespace

# Application Backup and Restore API

This repository contains APIs for managing application backups and restores using HTTP requests. It provides functionality to register applications, perform backups, and restore applications.

## APIs

### Register Application

Registers an application in the system.

**Endpoint:** `PUT /application/`

**Request Body:**
```json
{
    "namespace": "test-mariadb",
    "name": "mariadb"
}
```

**Response:**
```json
{
    "app_id": "app_1"
}
```

### Backup Application

Initiates a backup for the registered application.

**Endpoint:** `PUT /backup/`

**Request Body:**
```json
{
    "app_id": "app_1"
}
```

**Response:**
```json
{
    "app_id": "app_1",
    "backup_id": "backup_1"
}
```

### Restore Application

Restores a backed-up application.

**Endpoint:** `PUT /restore/`

**Request Body:**
```json
{
    "namespace": "demo9",
    "backup_id": "backup_3"
}
```

**Response:**
```json
{
    "message": "Restore completed successfully"
}
```

## How to Run Locally
To run the app locally, follow these steps:

Clone the repository using Git:
```bash
git clone https://github.com/kinhalkarrushikesh123/net_exercise.git
cd net_exercise
go run main.go
```
# or

```bash
git clone https://github.com/kinhalkarrushikesh123/net_exercise.git
cd net_exercise
go build -o backup
./backup
```

## Running using Docker
```bash
docker build -t backup:latest .
docker run -p -v (Path to KubeConfig):/root/.kube/config 8080:8080 backup:latest

```
Note:
Please ensure that the Docker container has access to your Kubernetes cluster's API server and has all the necessary permissions and privileges to interact with the API server. This includes appropriate authentication credentials, authorization for performing actions on Kubernetes resources, and network connectivity to reach the API server endpoints. Ensuring proper access and permissions is crucial for the Docker container to successfully communicate with the Kubernetes API server and perform desired operations within the cluster.