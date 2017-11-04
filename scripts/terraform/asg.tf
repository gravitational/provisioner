resource "aws_autoscaling_group" "nodes" {
  name                      = "${var.cluster_name}"
  max_size                  = 10
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "EC2"
  desired_capacity          = 0
  force_delete              = false
  launch_configuration      = "${aws_launch_configuration.node.name}"
  vpc_zone_identifier       = ["${aws_subnet.private.*.id}"]

  tags = "${local.asg_tags}"

  // external autoscale algos can modify these values,
  // so ignore changes to them
  lifecycle {
    ignore_changes = ["desired_capacity", "max_size", "min_size"]
  }
}

data "template_file" "node_user_data" {
  template = "${file("node-user-data.tpl")}"
  vars {
    cluster_name = "${var.cluster_name}"
  }
}

resource "aws_launch_configuration" "node" {
  name                        = "${var.cluster_name}"
  image_id                    = "${data.aws_ami.base.id}"
  instance_type               = "${var.node_instance_type}"
  user_data                   = "${data.template_file.node_user_data.rendered}"
  key_name                    = "${var.key_name}"
  ebs_optimized               = true
  associate_public_ip_address = false
  security_groups             = ["${aws_security_group.kubernetes.id}"]
  iam_instance_profile        = "${aws_iam_instance_profile.node.id}"

  root_block_device {
    delete_on_termination = true
    volume_type           = "io1"
    volume_size           = "50"
    iops                  = 500
  }

  ebs_block_device = {
    delete_on_termination = true
    volume_type           = "io1"
    volume_size           = "500"
    device_name           = "/dev/xvdb"
    iops                  = 500
  }
}

resource "aws_autoscaling_lifecycle_hook" "launching" {
  name                   = "${local.safe_cluster_name}-launching"
  autoscaling_group_name = "${aws_autoscaling_group.nodes.name}"
  default_result         = "CONTINUE"
  heartbeat_timeout      = 60
  lifecycle_transition   = "autoscaling:EC2_INSTANCE_LAUNCHING"

  notification_metadata = <<EOF
{
  "KubernetesCluster": "${var.cluster_name}"
}
EOF

  notification_target_arn = "${aws_sqs_queue.lifecycle_hooks.arn}"
  role_arn                = "${aws_iam_role.lifecycle_hooks.arn}"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_lifecycle_hook" "terminating" {
  name                   = "${local.safe_cluster_name}-terminating"
  autoscaling_group_name = "${aws_autoscaling_group.nodes.name}"
  default_result         = "CONTINUE"
  heartbeat_timeout      = 60
  lifecycle_transition   = "autoscaling:EC2_INSTANCE_TERMINATING"

  notification_metadata = <<EOF
{
  "KubernetesCluster": "${var.cluster_name}"
}
EOF

  notification_target_arn = "${aws_sqs_queue.lifecycle_hooks.arn}"
  role_arn                = "${aws_iam_role.lifecycle_hooks.arn}"

  lifecycle {
    create_before_destroy = true
  }
}
