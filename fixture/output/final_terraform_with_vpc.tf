// these variables are loaded from
// existing AWS environment and VPC

// https://www.terraform.io/docs/providers/aws/d/availability_zones.html
variable "azs" {
   default = ["us-west-1", "us-west-2", "us-west-3"]
}

// https://www.terraform.io/docs/providers/aws/d/subnet_ids.html
variable "subnets" {
   default = ["10.0.3.0/24", "10.0.4.0/24", "10.0.5.0/24"]
}

variable "public_subnets" {
   default = ["10.0.6.0/24", "10.0.7.0/24", "10.0.8.0/24"]
}

variable "aws_region" {
  default = "us-west"
}

data "aws_internet_gateway" "default" {
  filter {
    name = "attachment.vpc-id"
    values = ["vpc-1"]
  }

  filter {
    name = "attachment.state"
    values = ["available"]
  }
}

locals {
  vpc_id = "vpc-1"
  internet_gateway_id = "${data.aws_internet_gateway.default.id}"
  nat_gateways = ["ngw1", "ngw2", "ngw3"]
}
