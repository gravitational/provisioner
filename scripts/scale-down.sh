#!/bin/bash
set -e
set -x

#
# Scale down the cluster, by finding the aws id of the requested instance and whether it belong to an asg.
# if it's in an asg, remove it and decrease the desired-capacity, if it isn't just remove the node
#


# Find the existing node
nodeToDelete=`aws --region ${AWS_REGION} --profile default ec2 describe-instances --filters "Name=private-ip-address,Values=${AWS_INSTANCE_PRIVATE_IP}" "Name=tag:Name,Values=${TELEKUBE_CLUSTER_NAME}" --query "Reservations[0].Instances[0].InstanceId" | sed 's/\"//g'`
echo "nodeToDelete: " ${nodeToDelete}

if [ -z "$nodeToDelete" ]; then
  echo "unable to find node for ip: " ${AWS_INSTANCE_PRIVATE_IP}
  exit 1
fi

# Find out if it's a member of an autoscaling group
scalinggroup=`aws --region ${AWS_REGION} --profile default autoscaling describe-auto-scaling-instances --instance-ids ${nodeToDelete} --query "AutoScalingInstances[0].AutoScalingGroupName"`
echo "scalinggroup: " ${scalinggroup}

if [ "${scalinggroup}" = "null" ]; then
  aws --region ${AWS_REGION} --profile default ec2 terminate-instances --instance-ids ${nodeToDelete}
else
  aws --region ${AWS_REGION} --profile default autoscaling terminate-instance-in-auto-scaling-group --instance-id ${nodeToDelete} --should-decrement-desired-capacity
fi
