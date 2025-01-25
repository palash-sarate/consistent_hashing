# Start Minikube
if (-not (minikube status --format '{{.Host}}' | Select-String 'Running')) {
    minikube start
}
# build the docker image
docker build -t cache-service:latest .
# delete the cache-service deployment
# kubectl delete deployment cache-service
# # add the docker image to minikube
# minikube image rm cache-service:latest
# minikube image load cache-service:latest
# # Initialize Terraform
# terraform init
# # Apply Terraform configuration
# terraform apply -auto-approve