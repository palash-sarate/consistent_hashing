provider "kubernetes" {
  config_path = "~/.kube/config"
}

resource "kubernetes_namespace" "ch-demo" {
  metadata {
    name = "ch-demo"
  }
}

resource "kubernetes_deployment" "cache_service" {
  metadata {
    name      = "cache-service"
    namespace = kubernetes_namespace.ch-demo.metadata[0].name
  }

  spec {
    replicas = 3

    selector {
      match_labels = {
        app = "cache-service"
      }
    }

    template {
      metadata {
        labels = {
          app = "cache-service"
        }
      }

      spec {
        container {
          image = "cache-service:latest"
          name  = "cache-service"

          image_pull_policy = "IfNotPresent"

          port {
            container_port = 8080
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "cache_service" {
  metadata {
    name      = "cache-service"
    namespace = kubernetes_namespace.ch-demo.metadata[0].name
  }

  spec {
    selector = {
      app = "cache-service"
    }

    port {
      port        = 80
      target_port = 8080
    }

    type = "LoadBalancer"
  }
}