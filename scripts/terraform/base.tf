terraform {
  required_version = "~> 0.10.8"
  backend "s3" {}
}

provider "random" {
  version = "~> 1.0"
}

provider "template" {
  version = "~> 1.0"
}

data "aws_region" "current" {
  current = true
}

data "aws_caller_identity" "current" {}
