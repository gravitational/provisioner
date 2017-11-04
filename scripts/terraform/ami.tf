data "aws_ami" "base" {
  most_recent = true
  owners      = ["self"]

  filter {
    name   = "name"
    values = ["centos-7-k8s-base-ami *"]
  }
}
