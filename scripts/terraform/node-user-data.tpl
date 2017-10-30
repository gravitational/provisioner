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

# install python to get access to SSM
curl -O https://bootstrap.pypa.io/get-pip.py
python2.7 get-pip.py
pip install awscli
EC2_AVAIL_ZONE=`curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone`
EC2_REGION="`echo \"$EC2_AVAIL_ZONE\" | sed -e 's:\([0-9][0-9]*\)[a-z]*\$:\\1:'`"
TELEKUBE_SERVICE=`aws ssm get-parameter --name /telekube/${cluster_name}/service --region $EC2_REGION --output text 2>&1 | awk '{ print $4 }'`

# Download gravity of the right version directly from the cluster
curl -k -o /tmp/gravity $${TELEKUBE_SERVICE}/telekube/gravity
chmod +x /tmp/gravity

# In AWS mode gravity will discover the data from AWS SSM
/tmp/gravity discover ${cluster_name} --role=knode
