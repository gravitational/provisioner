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

# This calls opscenter to start the provision k8s job
curl --tlsv1.2 --insecure '${ops_url}/${ops_token}/node?provisioner=aws_terraform&bg=true' | bash
