#!/bin/bash
set -e
set -x

#
# Scale up the cluster, by finding the current desired capacity of the cluster auto-scaling group
# and then increasing the desired capacity by 1 server
#

ls -l $HOME/.aws
cat $HOME/.aws/credentials

currentCapacity=`aws --region ${AWS_REGION} --profile default autoscaling describe-auto-scaling-groups --auto-scaling-group-names ${TELEKUBE_CLUSTER_NAME} --query "AutoScalingGroups[0].DesiredCapacity"`
echo "current: " ${currentCapacity}
desiredCapacity=`expr ${currentCapacity} + ${TELEKUBE_NODE_PROFILE_ADD_COUNT_node}`
echo "desired: " ${desiredCapacity}
aws --region ${AWS_REGION} --profile default autoscaling set-desired-capacity --auto-scaling-group-name ${TELEKUBE_CLUSTER_NAME} --desired-capacity ${desiredCapacity}
