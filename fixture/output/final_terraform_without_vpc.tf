// these variables are loaded from
// existing AWS environment and VPC
data "aws_ami" "base" {
  most_recent = true
  owners = ["self"]
  filter {
    name = "name"
    values = ["centos-7-k8s-base-ami *"]
  }
}

provider "aws" {
  region      = "us-west-1"
  shared_credentials_file = "/var/lib/telekube/aws-credentials"
  profile                 = "default"
  max_retries = 5
}

variable "vpc_cidr" {
  default = "10.1.0.0/16"
}

variable "vpc_id" {
  default = ""
}

variable "internet_gateway_id" {
  default = ""
}

// Generate a same length array with azs. We don't need this value when
// creating VPC. However, terraform still needs a reference to the variable
// even the value is not being used so we set it to an array of empty
// string
variable "nat_gateways" {
  default = ["", "", ""]
}

variable "azs" {
   default = ["us-west-1", "us-west-2", "us-west-3"]
}

variable "subnets" {
   default = ["10.1.0.0/24", "10.1.2.0/24", "10.1.4.0/24"]
}

variable "public_subnets" {
   default = ["10.1.1.0/24", "10.1.3.0/24", "10.1.5.0/24"]
}


resource "aws_vpc" "kubernetes" {
  cidr_block            = "10.1.0.0/16"
  enable_dns_support    = true
  enable_dns_hostnames  = true

  tags {
    KubernetesCluster = "${var.cluster_name}"
    Name              = "${var.cluster_name}"
  }
}


resource "aws_eip" "nat" {
  count = "${length(var.azs)}"
  vpc   = true
}


resource "aws_internet_gateway" "kubernetes" {
  vpc_id = "${aws_vpc.kubernetes.id}"

  tags {
    KubernetesCluster = "${var.cluster_name}"
    Name              = "${var.cluster_name}"
  }
}


resource "aws_nat_gateway" "kubernetes" {
  count         = "${length(var.azs)}"
  allocation_id = "${element(aws_eip.nat.*.id, count.index)}"
  subnet_id     = "${element(aws_subnet.public.*.id, count.index)}"
  depends_on = ["aws_subnet.public", "aws_internet_gateway.kubernetes"]
}

vpc_id = "${aws_vpc.kubernetes.id}"
