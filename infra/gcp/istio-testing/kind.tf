resource "docker_image" "kindest-kind-node-amd64" {
  for_each = local.kind_node_images

  name         = each.key
  keep_locally = false
  platform     = "linux/amd64"
}

resource "docker_image" "kindest-kind-node-arm64" {
  for_each = local.kind_node_images

  name         = each.key
  keep_locally = false
  platform     = "linux/arm64"
}

resource "docker_tag" "kind-node-retag" {
  for_each = local.kind_node_images

  source_image = each.key
  target_image = "gcr.io/istio-testing/kind-node:${each.value}"
  depends_on = [
    docker_image.kindest-kind-node-amd64,
    docker_image.kindest-kind-node-arm64,
  ]
}

resource "docker_registry_image" "kind-node" {
  for_each = local.kind_node_images

  name          = "gcr.io/istio-testing/kind-node:${each.value}"
  keep_remotely = true # Don't remove the remote image on a terraform delete

  depends_on = [
    docker_tag.kind-node-retag,
    docker_image.kindest-kind-node-amd64,
    docker_image.kindest-kind-node-arm64,
  ]
}
