# docker-bake.hcl
group "default" {
  targets = ["srt-dashboard", "srt-consumer"]
}

variable "GIT_TAG" {
  default = "latest"
}

variable "REGISTRY" {
  default = "ghcr.io"
}

variable "REPO" {
  default = "phillezi/server-room-temperature"
}

target "common" {
  cache-from = ["type=gha"]
  cache-to   = ["type=gha,mode=max"]

  args = {
    GIT_TAG = "${GIT_TAG}"
  }

  # platforms = ["linux/amd64", "linux/arm64"]
}

target "srt-dashboard" {
  inherits = ["common"]
  dockerfile = "docker/srt-dashboard/Dockerfile"

  tags = concat(
    ["${REGISTRY}/${REPO}/srt-dashboard:${GIT_TAG}"],
    GIT_TAG == "latest" ? [] : ["${REGISTRY}/${REPO}/srt-dashboard:latest"]
  )
}

target "srt-consumer" {
  inherits = ["common"]
  dockerfile = "docker/srt-consumer/Dockerfile"

  tags = concat(
    ["${REGISTRY}/${REPO}/srt-consumer:${GIT_TAG}"],
    GIT_TAG == "latest" ? [] : ["${REGISTRY}/${REPO}/srt-consumer:latest"]
  )
}
