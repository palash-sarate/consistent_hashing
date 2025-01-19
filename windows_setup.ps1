# Start Minikube
minikube start

# Set the context to Minikube
kubectl config use-context minikube

# Verify the context
$currentContext = kubectl config current-context
Write-Output "Current context: $currentContext"

# build the docker image
docker build -t cache-service:latest .

# add the docker image to minikube
minikube image load cache-service:latest

# Initialize Terraform
terraform init

# Apply Terraform configuration
terraform apply -auto-approve

