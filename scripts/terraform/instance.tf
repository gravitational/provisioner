data "template_file" "user_data" {
  template = "${file("user-data.tpl")}"

  vars {
    ops_url   = "${var.ops_url}"
    ops_token = "${var.ops_token}"
  }
}

resource "aws_instance" "master" {
  count                       = "${var.node_count}"
  ami                         = "${data.aws_ami.base.id}"
  instance_type               = "${var.node_instance_type}"
  key_name                    = "${var.key_name}"
  associate_public_ip_address = false
  source_dest_check           = false
  vpc_security_group_ids      = ["${aws_security_group.kubernetes.id}"]
  subnet_id                   = "${element(aws_subnet.private.*.id, count.index % length(aws_subnet.private.*.id))}"
  iam_instance_profile        = "${aws_iam_instance_profile.master.id}"
  ebs_optimized               = true
  user_data                   = "${data.template_file.user_data.rendered}"
  tags                        = "${merge(local.common_tags, map())}"
  volume_tags                 = "${merge(local.common_tags, map())}"

  lifecycle {
    ignore_changes = ["user_data", "instance_type"]
  }

  root_block_device {
    delete_on_termination = true
    volume_type           = "io1"
    volume_size           = "50"
    iops                  = 500
  }

  # /var/lib/gravity device with all the stuff (docker, etc)
  ebs_block_device = {
    delete_on_termination = true
    volume_type           = "io1"
    volume_size           = "500"
    device_name           = "/dev/xvdb"
    iops                  = 1500
  }

  # etcd device on a separate disk, so it's not too flaky
  ebs_block_device = {
    delete_on_termination = true
    volume_type           = "io1"
    volume_size           = "100"
    device_name           = "/dev/xvdf"
    iops                  = 1500
  }
}

resource "aws_iam_instance_profile" "master" {
  name = "master-${var.cluster_name}"
  role = "${aws_iam_role.master.name}"
}

resource "aws_iam_instance_profile" "node" {
  name = "node-${var.cluster_name}"
  role = "${aws_iam_role.node.name}"
}
