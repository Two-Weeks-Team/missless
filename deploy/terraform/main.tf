# ============================================================
# missless.co — Terraform Infrastructure
# Google Cloud Run + Firestore + Storage + IAM
# ============================================================

terraform {
  required_version = ">= 1.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# ── Required APIs ────────────────────────────────────────────

resource "google_project_service" "run" {
  service            = "run.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "artifactregistry" {
  service            = "artifactregistry.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "firestore" {
  service            = "firestore.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "secretmanager" {
  service            = "secretmanager.googleapis.com"
  disable_on_destroy = false
}

# ── Artifact Registry ────────────────────────────────────────

resource "google_artifact_registry_repository" "missless" {
  location      = var.region
  repository_id = "missless"
  format        = "DOCKER"
  description   = "Docker images for missless.co backend"

  depends_on = [google_project_service.artifactregistry]
}

# ── Service Account ──────────────────────────────────────────

resource "google_service_account" "backend" {
  account_id   = "missless-backend"
  display_name = "missless Backend Service Account"
  description  = "Service account for missless.co Cloud Run backend"
}

# ── IAM Bindings ─────────────────────────────────────────────

resource "google_project_iam_member" "vertex_ai_user" {
  project = var.project_id
  role    = "roles/aiplatform.user"
  member  = "serviceAccount:${google_service_account.backend.email}"
}

resource "google_project_iam_member" "firestore_user" {
  project = var.project_id
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.backend.email}"
}

resource "google_project_iam_member" "storage_admin" {
  project = var.project_id
  role    = "roles/storage.objectAdmin"
  member  = "serviceAccount:${google_service_account.backend.email}"
}

resource "google_project_iam_member" "secret_accessor" {
  project = var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_service_account.backend.email}"
}

# ── Cloud Firestore ──────────────────────────────────────────

resource "google_firestore_database" "default" {
  project     = var.project_id
  name        = "(default)"
  location_id = var.region
  type        = "FIRESTORE_NATIVE"

  depends_on = [google_project_service.firestore]
}

# ── Cloud Storage ────────────────────────────────────────────

resource "google_storage_bucket" "assets" {
  name          = "${var.project_id}-assets"
  location      = var.region
  storage_class = "STANDARD"

  uniform_bucket_level_access = true

  cors {
    origin          = ["https://missless.co", "http://localhost:18080"]
    method          = ["GET", "HEAD"]
    response_header = ["Content-Type"]
    max_age_seconds = 3600
  }

  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
  }
}

# Public access for albums
resource "google_storage_bucket_iam_member" "albums_public" {
  bucket = google_storage_bucket.assets.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"

  condition {
    title      = "albums_only"
    expression = "resource.name.startsWith('projects/_/buckets/${google_storage_bucket.assets.name}/objects/albums/')"
  }
}

# ── Cloud Run Service ────────────────────────────────────────

resource "google_cloud_run_v2_service" "missless" {
  name     = "missless"
  location = var.region

  depends_on = [google_project_service.run]

  template {
    service_account = google_service_account.backend.email

    scaling {
      min_instance_count = 1
      max_instance_count = 10
    }

    session_affinity = true

    containers {
      image = "asia-northeast3-docker.pkg.dev/${var.project_id}/missless/server:latest"

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      env {
        name  = "GCP_PROJECT_ID"
        value = var.project_id
      }
      env {
        name  = "ENVIRONMENT"
        value = "production"
      }
      env {
        name  = "DOMAIN"
        value = "missless.co"
      }
      env {
        name  = "STORAGE_BUCKET"
        value = google_storage_bucket.assets.name
      }
    }

    timeout = "3600s"
  }
}

# Allow unauthenticated access
resource "google_cloud_run_v2_service_iam_member" "public" {
  project  = var.project_id
  location = var.region
  name     = google_cloud_run_v2_service.missless.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# ── Domain Mapping ───────────────────────────────────────────

resource "google_cloud_run_domain_mapping" "missless" {
  location = var.region
  name     = "missless.co"

  metadata {
    namespace = var.project_id
  }

  spec {
    route_name = google_cloud_run_v2_service.missless.name
  }
}
