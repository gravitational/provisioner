// these variables are loaded from
// existing AWS environment and VPC

variable "azs" {
   default = ["us-west-1", "us-west-2", "us-west-3"]
}

variable "subnets" {
   default = ["10.0.3.0/24", "10.0.4.0/24", "10.0.5.0/24"]
}

variable "public_subnets" {
   default = ["10.0.6.0/24", "10.0.7.0/24", "10.0.8.0/24"]
}

variable "aws_region" {
  default = "us-west"
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
