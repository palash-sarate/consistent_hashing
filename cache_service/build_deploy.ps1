# Start Minikube
if (-not (minikube status --format '{{.Host}}' | Select-String 'Running')) {
    minikube start
}
# build the docker image
docker build -t cache-service:latest .
# delete the cache-node deployment
try {
    kubectl delete deployment cache-node -n ch-demo | Out-Null
    kubectl delete deployment cache-service -n ch-demo | Out-Null
    kubectl delete service cache-service -n ch-demo | Out-Null
} catch {
    Write-Host "Failed to delete deployment cache-node and service: $_"
}
# add the docker image to minikube
minikube image rm cache-service:latest
minikube image load cache-service:latest
# Navigate to the root directory where the main.tf file is located
Push-Location -Path ".."

# Initialize Terraform
terraform init

# Apply Terraform configuration
terraform apply -auto-approve

# Return to the original directory
Push-Location -Path "./cache_service"