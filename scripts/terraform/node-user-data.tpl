#!/bin/bash
set -x

# Set some curl options so that temporary failures get retried
# More info: https://ec.haxx.se/usingcurl-timeouts.html
CURL_OPTS="--retry 100 --retry-delay 0 --connect-timeout 10 --max-time 300"

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

# install python to get access to SSM
curl $${CURL_OPTS} -O https://bootstrap.pypa.io/get-pip.py
python2.7 get-pip.py
pip install awscli
EC2_AVAIL_ZONE=`curl $${CURL_OPTS} -s http://169.254.169.254/latest/meta-data/placement/availability-zone`
EC2_REGION="`echo \"$EC2_AVAIL_ZONE\" | sed -e 's:\([0-9][0-9]*\)[a-z]*\$:\\1:'`"
TELEKUBE_SERVICE=`aws ssm get-parameter --name /telekube/${cluster_name}/service --region $EC2_REGION --query 'Parameter.Value' --output text`

# Explicitly configure required parameters
modprobe iptable_nat || true
modprobe iptable_filter || true
modprobe overlay || true
modprobe br_netfilter || true
modprobe ebtable_filter || true
cat > /etc/sysctl.d/50-telekube.conf <<EOF
net.ipv4.ip_forward=1
net.bridge.bridge-nf-call-iptables=1
EOF
if sysctl -q fs.may_detach_mounts >/dev/null 2>&1; then
  echo "fs.may_detach_mounts=1" >> /etc/sysctl.d/50-telekube.conf
fi

cat > /etc/modules-load.d/telekube.conf <<EOF
iptable_nat
iptable_filter
br_netfilter
overlay
ebtable_filter
EOF
sysctl -p /etc/sysctl.d/50-telekube.conf

# Download gravity of the right version directly from the cluster
curl $${CURL_OPTS} -k -o /tmp/gravity $${TELEKUBE_SERVICE}/telekube/gravity
chmod +x /tmp/gravity

# In AWS mode gravity will discover the data from AWS SSM and join the cluster
/tmp/gravity autojoin ${cluster_name} --role=knode
