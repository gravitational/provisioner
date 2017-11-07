// private subnets and routing tables
resource "aws_route_table" "private" {
  count  = "${length(var.azs)}"
  vpc_id = "${local.vpc_id}"

  tags = "${merge(local.common_tags, map(
    "Name", "${var.cluster_name}-private"
  ))}"
}

resource "aws_route" "private_nat" {
  count                  = "${length(var.azs)}"
  route_table_id         = "${element(aws_route_table.private.*.id, count.index)}"
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = "${element(local.nat_gateways, count.index)}"
  depends_on             = ["aws_route_table.private"]
}

resource "aws_subnet" "private" {
  count             = "${length(var.azs)}"
  vpc_id            = "${local.vpc_id}"
  cidr_block        = "${element(var.subnets, count.index)}"
  availability_zone = "${element(var.azs, count.index)}"

  tags = "${merge(local.common_tags, map(
    "Name", "${var.cluster_name}-private",
    "kubernetes.io/role/internal-elb", "1"
  ))}"
}

resource "aws_route_table_association" "private" {
  count          = "${length(var.azs)}"
  subnet_id      = "${element(aws_subnet.private.*.id, count.index)}"
  route_table_id = "${element(aws_route_table.private.*.id, count.index)}"
}

// public subnets and routing tables
resource "aws_route_table" "public" {
  count  = "${length(var.azs)}"
  vpc_id = "${local.vpc_id}"

  tags = "${merge(local.common_tags, map(
    "Name", "${var.cluster_name}-public"
  ))}"
}

resource "aws_route" "public_gateway" {
  count                  = "${length(var.azs)}"
  route_table_id         = "${element(aws_route_table.public.*.id, count.index)}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${local.internet_gateway_id}"
  depends_on             = ["aws_route_table.public"]
}

resource "aws_subnet" "public" {
  count             = "${length(var.azs)}"
  vpc_id            = "${local.vpc_id}"
  cidr_block        = "${element(var.public_subnets, count.index)}"
  availability_zone = "${element(var.azs, count.index)}"

  tags = "${merge(local.common_tags, map(
    "Name", "${var.cluster_name}-public",
    "kubernetes.io/role/elb", "1"
  ))}"
}

resource "aws_route_table_association" "public" {
  count          = "${length(var.azs)}"
  subnet_id      = "${element(aws_subnet.public.*.id, count.index)}"
  route_table_id = "${element(aws_route_table.public.*.id, count.index)}"
}
