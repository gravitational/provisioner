resource "aws_instance" "node" {
  count                       = "${var.node_count}"
  ami                         = "${data.aws_ami.base.id}"
  instance_type               = "${var.node_instance_type}"
  key_name                    = "${var.key_name}"
  associate_public_ip_address = false
  source_dest_check           = false
  vpc_security_group_ids      = ["${aws_security_group.kubernetes.id}"]
  subnet_id                   = "${element(aws_subnet.private.*.id, count.index % length(aws_subnet.private.*.id))}"
  iam_instance_profile        = "${aws_iam_instance_profile.node.id}"
  ebs_optimized               = true
  lifecycle {
     ignore_changes = [ "user_data", "instance_type" ]
  }

  user_data = <<EOF
#!/bin/bash
set -x

umount /dev/xvdb
mkfs.ext4 /dev/xvdb
sed -i.bak '/xvdb/d' /etc/fstab
echo -e '/dev/xvdb\t/var/lib/gravity\text4\tdefaults\t0\t2' >> /etc/fstab
mkdir -p /var/lib/gravity
mount /var/lib/gravity
chown -R 1000:1000 /var/lib/gravity
sed -i.bak 's/Defaults    requiretty/#Defaults    requiretty/g' /etc/sudoers
export SUDO_USER=centos
export SUDO_UID=1000
export SUDO_GID=1000

curl --tlsv1.2 --insecure '${var.ops_url}/${var.ops_token}/node?provisioner=aws_terraform&bg=true' | bash

EOF

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
    iops                  = 1500
  }

  tags {
    KubernetesCluster = "${var.cluster_name}"
   }
}

resource "aws_iam_instance_profile" "node" {
  name       = "node-${var.cluster_name}"
  role       = "${aws_iam_role.master.name}"
  depends_on = ["aws_iam_role_policy.master"]
  provisioner "local-exec" {
    command = "sleep 30"
  }
}