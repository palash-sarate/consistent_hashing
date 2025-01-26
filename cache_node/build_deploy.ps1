# Start Minikube
if (-not (minikube status --format '{{.Host}}' | Select-String 'Running')) {
    minikube start
}
# build the docker image
docker build -t cache-node:latest .
# delete the cache-node deployment
try {
    kubectl delete deployment cache-node -n ch-demo
    kubectl delete deployment cache-service -n ch-demo
} catch {
    Write-Host "Failed to delete deployment cache-node and service: $_"
}
# add the docker image to minikube
minikube image rm cache-node:latest
minikube image load cache-node:latest

# Navigate to the root directory where the main.tf file is located
Push-Location -Path ".."

# Initialize Terraform
terraform init

# Apply Terraform configuration
terraform apply -auto-approve

# Return to the original directory
Pop-Location