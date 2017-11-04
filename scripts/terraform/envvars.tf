// these variables are supplied via environment
// during the provisioning

// SSH key name
variable "key_name" {}

// cluster name
variable "cluster_name" {}

// ops center URL
variable "ops_url" {}

// ops center token
variable "ops_token" {}

// amount of nodes to set up
variable "node_count" {}

// instance types for nodes
variable "node_instance_type" {}

// AWS KMS alias used for encryption/decryption
// default is alias used in SSM
variable "kms_alias_name" {
  default = "alias/aws/ssm"
}

// safe cluster name to use in places sensitive to naming, e.g. SQS queues and lifecycle hooks
locals {
  safe_cluster_name = "${replace(var.cluster_name, "/[^a-zA-Z0-9\\-]/", "")}"
}

// common tags required on all resources
locals {
  common_tags = {
    Terraform         = "true"
    KubernetesCluster = "${var.cluster_name}"
    Name              = "${var.cluster_name}"
  }
}

// Create ASG tag setup from common tags
resource "null_resource" "asg_tags" {
  count = "${length(local.common_tags)}"
  triggers {
    key                 = "${element(keys(local.common_tags), count.index)}",
    value               = "${element(values(local.common_tags), count.index)}",
    propagate_at_launch = true
  }
}

locals {
    asg_tags = ["${null_resource.asg_tags.*.triggers}"]
}
