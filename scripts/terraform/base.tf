terraform {
  required_version = "~> 0.11.3"
  backend "s3" {}
}

provider "random" {
  version = "~> 1.0"
}

provider "template" {
  version = "~> 1.0"
}

variable "aws_max_retries" {
  default = 5
}

provider "aws" {
  version                 = "~> 1.13"
  region                  = "${var.aws_region}"
  shared_credentials_file = "/var/lib/telekube/aws-credentials"
  profile                 = "default"
  max_retries             = "${var.aws_max_retries}"
}

data "aws_region" "current" {
  current = true
}

data "aws_caller_identity" "current" {}
