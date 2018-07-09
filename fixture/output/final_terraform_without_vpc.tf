// these variables are loaded from
// existing AWS environment and VPC

// https://www.terraform.io/docs/providers/aws/d/availability_zones.html
variable "azs" {
   default = ["us-west-1", "us-west-2", "us-west-3"]
}

// https://www.terraform.io/docs/providers/aws/d/subnet_ids.html
variable "subnets" {
   default = ["10.1.0.0/24", "10.1.2.0/24", "10.1.4.0/24"]
}

variable "public_subnets" {
   default = ["10.1.1.0/24", "10.1.3.0/24", "10.1.5.0/24"]
}

variable "aws_region" {
  default = "us-west-1"
}

variable "vpc_cidr" {
  default = "10.1.0.0/16"
}

resource "aws_vpc" "kubernetes" {
  cidr_block            = "10.1.0.0/16"
  enable_dns_support    = true
  enable_dns_hostnames  = true
  tags                  = "${merge(local.common_tags, map())}"
}

resource "aws_eip" "nat" {
  count = "${length(var.azs)}"
  vpc   = true
}

resource "aws_internet_gateway" "kubernetes" {
  vpc_id = "${aws_vpc.kubernetes.id}"
  tags   = "${merge(local.common_tags, map())}"
}

resource "aws_nat_gateway" "kubernetes" {
  count         = "${length(var.azs)}"
  allocation_id = "${element(aws_eip.nat.*.id, count.index)}"
  subnet_id     = "${element(aws_subnet.public.*.id, count.index)}"
  tags          = "${merge(local.common_tags, map())}"
  depends_on    = ["aws_subnet.public", "aws_internet_gateway.kubernetes"]
}

locals {
  vpc_id = "${aws_vpc.kubernetes.id}"
  internet_gateway_id = "${aws_internet_gateway.kubernetes.id}"
  nat_gateways = ["${aws_nat_gateway.kubernetes.*.id}"]
}
