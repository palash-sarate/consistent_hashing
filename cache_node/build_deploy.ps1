# Start Minikube
minikube start
# build the docker image
docker build -t cache-node:latest .
# delete the cache-node deployment
kubectl delete deployment cache-node
# add the docker image to minikube
minikube image rm cache-node:latest
minikube image load cache-node:latest
# Initialize Terraform
terraform init
# Apply Terraform configuration
terraform apply -auto-approve