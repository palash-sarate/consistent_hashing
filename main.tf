provider "kubernetes" {
  config_path = "~/.kube/config"
}

resource "kubernetes_namespace" "ch-demo" {
  metadata {
    name = "ch-demo"
  }
}

resource "kubernetes_service_account" "cache_service_sa" {
  metadata {
    name      = "cache-service-sa"
    namespace = kubernetes_namespace.ch-demo.metadata[0].name
  }
}

resource "kubernetes_role" "cache_service_role" {
  metadata {
    name      = "cache-service-role"
    namespace = kubernetes_namespace.ch-demo.metadata[0].name
  }

  rule {
    api_groups = [""]
    resources  = ["pods"]
    verbs      = ["list"]
  }
}

resource "kubernetes_role_binding" "cache_service_role_binding" {
  metadata {
    name      = "cache-service-role-binding"
    namespace = kubernetes_namespace.ch-demo.metadata[0].name
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.cache_service_role.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.cache_service_sa.metadata[0].name
    namespace = kubernetes_namespace.ch-demo.metadata[0].name
  }
}
resource "kubernetes_deployment" "cache_service" {
  metadata {
    name      = "cache-service"
    namespace = kubernetes_namespace.ch-demo.metadata[0].name
  }

  spec {
    replicas = 1

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
        service_account_name = kubernetes_service_account.cache_service_sa.metadata[0].name
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

resource "kubernetes_deployment" "cache_nodes" {
  metadata {
    name      = "cache-node"
    namespace = kubernetes_namespace.ch-demo.metadata[0].name
  }

  spec {
    replicas = 3

    selector {
      match_labels = {
        app = "cache-node"
      }
    }

    template {
      metadata {
        labels = {
          app = "cache-node"
        }
      }

      spec {
        container {
          image = "cache-node:latest"
          name  = "cache-node"

          image_pull_policy = "IfNotPresent"

          port {
            container_port = 8080
          }
        }
      }
    }
  }
}

# resource "kubernetes_service" "cache_node" {
#   metadata {
#     name      = "cache-node"
#     namespace = kubernetes_namespace.ch-demo.metadata[0].name
#   }

#   spec {
#     selector = {
#       app = "cache-node"
#     }

#     port {
#       port        = 80
#       target_port = 8080
#     }

#     type = "LoadBalancer"
#   }
# }