## This is the script that is being executed
## by telekube provisioning hook

# cloud provider name
CLOUD_PROVIDER ?=
# AWS region anme
AWS_REGION ?=
# cloud image name
AWS_AMI  ?=
# AWS VPC ID
AWS_VPC_ID ?=
# AWS Key Name
AWS_KEY_NAME ?=
# AWS access key (for terraform)
AWS_ACCESS_KEY_ID ?=
# AWS secret key (for terraform)
AWS_SECRET_ACCESS_KEY ?=
# AWS session token (for terraform)
AWS_SESSION_TOKEN ?=
# variable name for telekube cluster
TELEKUBE_CLUSTER_NAME ?=
# telekube Ops Center URL
TELEKUBE_OPS_URL ?=
# Ops Center Token
TELEKUBE_OPS_TOKEN ?= $(shell cat /var/lib/telekube/token)
# application flavor set by user via UI or API
TELEKUBE_FLAVOR ?=
# All profile names
TELEKUBE_NODE_PROFILES ?=
# Telekube node profiles count requested via UI or API
# set to 0 if not set, otherwise new terraform will crash
TELEKUBE_NODE_PROFILE_COUNT_node ?= 0
# Telekube instance type
TELEKUBE_NODE_PROFILE_INSTANCE_TYPE_node ?=
# AWS instance private IP address of the instance to delete
AWS_INSTANCE_PRIVATE_IP ?=
# AWS instance private DNS name of the instance to delet
AWS_INSTANCE_PRIVATE_DNS ?=
# S3 bucket name to keep the state
BUCKET_NAME ?= provisioner-terraform-state
# directory with this makefile
CWD := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

# TF options
TF_FLAGS ?= -input -no-color

# export terraform variables
TF_VAR_bucket_name := $(BUCKET_NAME)
TF_VAR_aws_region ?= $(AWS_REGION)
TF_VAR_cluster_name ?= $(TELEKUBE_CLUSTER_NAME)
TF_VAR_key_name ?= $(AWS_KEY_NAME)
TF_VAR_ops_token ?= $(TELEKUBE_OPS_TOKEN)
TF_VAR_ops_url ?= $(TELEKUBE_OPS_URL)
TF_VAR_node_count ?= $(TELEKUBE_NODE_PROFILE_COUNT_node)
TF_VAR_node_instance_type ?= $(TELEKUBE_NODE_PROFILE_INSTANCE_TYPE_node)
export


# echo prints all variables used in excluding secrets
.PHONY: echo
echo:
	@echo "Cloud Provider:      $(CLOUD_PROVIDER)"
	@echo "Region:              $(AWS_REGION)"
	@echo "AMI:                 $(AWS_AMI)"
	@echo "VPC ID:              $(AWS_VPC_ID)"
	@echo "SSH key:             $(AWS_KEY_NAME)"
	@echo "Cluster Name:        $(TELEKUBE_CLUSTER_NAME)"
	@echo "Ops Center URL:      $(TELEKUBE_OPS_URL)"
	@echo "Ops Center Version:  $(TELEKUBE_OPS_VERSION)"
	@echo "App:                 $(TELEKUBE_APP)"
	@echo "Flavor:              $(TELEKUBE_FLAVOR)"
	@echo "Node profiles:       $(TELEKUBE_NODE_PROFILES)"
	@echo "Node count:          $(TELEKUBE_NODE_PROFILE_COUNT_node)"
	@echo "Node instance type:  $(TELEKUBE_NODE_PROFILE_INSTANCE_TYPE_node)"

#
# echo prints all variables excluding secrets
#
.PHONY: echo-nodes-deprovision
echo-nodes-deprovision:
	@echo "Cloud Provider:       $(CLOUD_PROVIDER)"
	@echo "Region:               $(AWS_REGION)"
	@echo "VPC ID:               $(AWS_VPC_ID)"
	@echo "Cluster Name:         $(TELEKUBE_CLUSTER_NAME)"
	@echo "Ops Center URL:       $(TELEKUBE_OPS_URL)"
	@echo "Ops Center Version:   $(TELEKUBE_OPS_VERSION)"
	@echo "App:                  $(TELEKUBE_APP)"
	@echo "Instance Private IP:  $(AWS_INSTANCE_PRIVATE_IP)"
	@echo "Instance Private DNS: $(AWS_INSTANCE_PRIVATE_DNS)"
	@echo "Node count:           $(TELEKUBE_NODE_PROFILE_COUNT_node)"

#
# clean-state cleans up temporary cluster state directory
#
.PHONY: clean-state
clean-state:
	rm -rf /mnt/state/cluster
	mkdir -p /mnt/state/cluster


# dry-run executes terraform plan without adding anything
.PHONY: dry-run
dry-run:
	$(MAKE) echo
	$(MAKE) fix-s3-backend
	$(MAKE) clean-state
	$(MAKE) init-cluster
	$(MAKE) sync-files
	$(MAKE) terraform-init
	$(MAKE) terraform-plan

#
# cluster-provision provisions cluster
#
.PHONY: cluster-provision
cluster-provision:
	$(MAKE) echo
	$(MAKE) fix-s3-backend
	$(MAKE) clean-state
	$(MAKE) init-cluster
	$(MAKE) sync-files
	$(MAKE) terraform-init
	$(MAKE) terraform-apply


#
# cluster-deprovision deprovisions cluster
#
.PHONY: cluster-deprovision
cluster-deprovision:
	$(MAKE) echo
	$(MAKE) fix-s3-backend
	$(MAKE) clean-state
	$(MAKE) sync-files
	$(MAKE) terraform-init
	$(MAKE) terraform-destroy


#
# nodes-deprovision deprovisions nodes
#
.PHONY: nodes-deprovision
nodes-deprovision:
	$(MAKE) fix-s3-backend
	SHELL=/bin/bash

	# Find the existing node
	$(eval nodeToDelete=`aws --region ${AWS_REGION} --profile default ec2 describe-instances --filters "Name=private-ip-address,Values=${AWS_INSTANCE_PRIVATE_IP}" "Name=tag:Name,Values=${TELEKUBE_CLUSTER_NAME}" --query "Reservations[0].Instances[0].InstanceId" | sed 's/\"//g'`)
	@echo "nodeToDelete: " ${nodeToDelete}
	# Find out if it's a member of an autoscaling group
	$(eval scalinggroup=$(shell aws --region ${AWS_REGION} --profile default autoscaling describe-auto-scaling-instances --instance-ids ${nodeToDelete} --query "AutoScalingInstances[0].AutoScalingGroupName"))
	@echo "scalinggroup: " ${scalinggroup}

ifeq (${scalinggroup},"null")
	aws --region ${AWS_REGION} ec2 terminate-instances --instance-ids ${nodeToDelete}
else
	aws --region ${AWS_REGION} autoscaling terminate-instance-in-auto-scaling-group --instance-id ${nodeToDelete} --should-decrement-desired-capacity
endif


#
# nodes-provision provisions nodes
#
.PHONY: nodes-provision
nodes-provision:
	$(MAKE) fix-s3-backend
	SHELL=/bin/bash

	$(eval currentCapacity=$(shell aws --debug --region ${AWS_REGION} --profile default autoscaling describe-auto-scaling-groups --auto-scaling-group-names ${TELEKUBE_CLUSTER_NAME} --query "AutoScalingGroups[0].DesiredCapacity"))
	@echo "current: " ${currentCapacity}
	$(eval desiredCapacity=`expr ${currentCapacity} + ${TELEKUBE_NODE_PROFILE_ADD_COUNT_node}` )
	@echo "desired: " ${desiredCapacity}
	aws --debug --region ${AWS_REGION} --profile default autoscaling set-desired-capacity --auto-scaling-group-name ${TELEKUBE_CLUSTER_NAME} --desired-capacity ${desiredCapacity}

# this is a workaround for bug https://github.com/hashicorp/terraform/issues/13589
.PHONY: fix-s3-backend
fix-s3-backend:
	mkdir -p $$HOME/.aws
	ln -sf /var/lib/telekube/aws-credentials $$HOME/.aws/credentials

.PHONY: init-cluster
init-cluster:
	inspect init-vars \
		--region=$(AWS_REGION) \
		--vpc-id=$(AWS_VPC_ID) \
		--cluster-bucket=$(BUCKET_NAME) \
		--template=$(CWD)/terraform/templates/vars.tf.template \
		--key=$(TELEKUBE_CLUSTER_NAME)/cluster/vars.tf


# sync-files syncs files from S3 to input directory
.PHONY: sync-files
sync-files:
	env
	inspect sync-files \
		--region=$(AWS_REGION) \
		--cluster-bucket=$(BUCKET_NAME) \
		--target=/mnt/state/cluster/input \
		--prefix=$(TELEKUBE_CLUSTER_NAME)/cluster


.PHONY: terraform-init
terraform-init:
# copy user-data templates to the input directory
	cp $(CWD)/terraform/*.tpl /mnt/state/cluster/input
# copy terraform files to the input directory
	cp $(CWD)/terraform/*.tf /mnt/state/cluster/input
# init the config using s3 backend
	cd /mnt/state/cluster && \
		terraform init \
			-backend=true \
			-backend-config='bucket=$(BUCKET_NAME)' \
			-backend-config="region=$(AWS_REGION)" \
			-backend-config='key=$(TELEKUBE_CLUSTER_NAME)/cluster.tfstate' \
			--from-module='/mnt/state/cluster/input'

.PHONY: terraform-remove-instance
terraform-remove-instance:
	cd /mnt/state/cluster && terraform show | inspect find-instance --private-ip=$(AWS_INSTANCE_PRIVATE_IP) > /mnt/state/cluster/instance-to-remove
	export RESOURCE_NAME=$$(cat /mnt/state/cluster/instance-to-remove) &&  \
	echo "going to remove instance: $$RESOURCE_NAME" && \
	cd /mnt/state/cluster && \
		terraform destroy $(TF_FLAGS) --force --target="$$RESOURCE_NAME"

#
# terraform-plan runs terraform plan on the cluster
#
.PHONY: terraform-plan
terraform-plan:
	cd '/mnt/state/cluster' && \
		terraform plan $(TF_FLAGS)

#
# terraform-apply updates the existing infrastructure
#
.PHONY: terraform-apply
terraform-apply:
	cd '/mnt/state/cluster' && \
		terraform apply $(TF_FLAGS) -auto-approve

#
# terraform-destroy destroys the provisioned cluster
#
.PHONY: terraform-destroy
terraform-destroy:
	cd '/mnt/state/cluster' && \
		terraform destroy $(TF_FLAGS) --force

#
# terraform-destroy-instance destroys one instance
#
.PHONY: terraform-destroy-instance
terraform-destroy-instance:
	cd '/mnt/state/cluster' && \
		terraform destroy $(TF_FLAGS) --target
