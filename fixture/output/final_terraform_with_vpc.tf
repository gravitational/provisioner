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
  region      = "us-west"
  shared_credentials_file = "/var/lib/telekube/aws-credentials"
  profile                 = "default"
  max_retries = 5
}

variable "vpc_id" {
  default = "vpc-1"
}

variable "internet_gateway_id" {
  default = "igw-1"
}

variable "nat_gateways" {
  default = ["ngw1", "ngw2", "ngw3"]
}

variable "azs" {
   default = ["us-west-1", "us-west-2", "us-west-3"]
}

variable "subnets" {
   default = ["10.0.3.0/24", "10.0.4.0/24", "10.0.5.0/24"]
}

variable "public_subnets" {
   default = ["10.0.6.0/24", "10.0.7.0/24", "10.0.8.0/24"]
}

vpc_id = "${var.vpc_id}"
