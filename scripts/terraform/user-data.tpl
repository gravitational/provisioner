#!/bin/bash
set -xeuo pipefail

# Set some curl options so that temporary failures get retried
# More info: https://ec.haxx.se/usingcurl-timeouts.html
CURL_OPTS="--retry 100 --retry-delay 0 --connect-timeout 10 --max-time 300"

# mount all required volumes
umount /dev/xvdb || true
mkfs.ext4 /dev/xvdb
mkfs.ext4 /dev/xvdf
sed -i.bak '/xvdb/d' /etc/fstab
echo -e '/dev/xvdb\t/var/lib/gravity\text4\tdefaults\t0\t2' >> /etc/fstab
echo -e '/dev/xvdf\t/var/lib/gravity/planet/etcd\text4\tdefaults\t0\t2' >> /etc/fstab

mkdir -p /var/lib/gravity
mount /var/lib/gravity
mkdir -p /var/lib/gravity/planet/etcd
mount /var/lib/gravity/planet/etcd
chown -R 1000:1000 /var/lib/gravity /var/lib/gravity/planet/etcd

# Explicitly configure required parameters
cat > /etc/sysctl.d/50-telekube.conf <<EOF
net.ipv4.ip_forward=1
net.bridge.bridge-nf-call-iptables=1
EOF
if sysctl -q fs.may_detach_mounts >/dev/null 2>&1; then
  echo "fs.may_detach_mounts=1" >> /etc/sysctl.d/50-telekube.conf
fi

cat > /etc/modules-load.d/telekube.conf <<EOF
br_netfilter
overlay
ebtable_filter
EOF

sysctl -p /etc/sysctl.d/50-telekube.conf
modprobe overlay || true
modprobe br_netfilter || true
modprobe ebtable_filter || true

# Fix up sudoers
sed -i.bak 's/Defaults    requiretty/#Defaults    requiretty/g' /etc/sudoers
export SUDO_USER=centos
export SUDO_UID=1000
export SUDO_GID=1000

# This calls opscenter to start the provision k8s job
curl $${CURL_OPTS} --tlsv1.2 --insecure '${ops_url}/${ops_token}/node?provisioner=aws_terraform&bg=true' | bash
