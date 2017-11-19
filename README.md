# Provisioner

Terraform based provisioners for Ops Center

# Overview

Provisioner includes two components:

- an executable binary to generate terraform script
  to provision cluster
- a docker image bundling the above binary and a makefile exposing a set
  of tasks to provision a cluster

The docker image is the final package to be deployed and executed as a
kubernetes job to create VPC. The docker entrypoint is set to **Make**
so you can invoke any make target by passing command when executing with
docker run.

Provisioner works by generating a Terraform script and then applying the
generated script to create or modify resources.
The script will also (re)sync the configuration with an S3 bucket
whenever it is called.

Provisioner can run standalone, however, its main usage is to be
integrated with Telekube application hook to create/reuse VPC and launch
instance into the VPC.

Provisioner works in two modes:

1. Using an existing VPC
2. Creating a new VPC

## Existing VPC

In this mode, provisioner reads VPC configuration and reuses NAT gateway,
internet gateway. The generated Terraform will contain new resources for
subnets, security groups. The number of private/public subnet pairs will be
equal to the number of NAT gateways in the existing VPC. The subnet CIDR is
calculated to not overlap with any existing subnets in the same VPC.

## Creating a new VPC

Here provisioner generates a Terraform script to create a VPC from scratch:

  - NAT gateway and its Elastic IP respectively
  - Internet gateway and its Elastic IP respectively
  - Subnets: public and private and their route tables respectively

Provisioner creates subnets in all availability zones in the given region and
runs a NAT gateway on each availability zone.

## Environment variables

Please refer to `scripts/Makefile` for a list of available environment variables.
Below are required variables for all tasks:

* AWS_REGION: aws region of S3 bucket to store Terraform script and
  state.
* AWS_ACCESS_KEY_ID: aws access key id
* AWS_SECRET_ACCESS_KEY: aws secret key
* TELEKUBE_CLUSTER_NAME: telekube cluster. We will store generated
  script in a bucket in form of **provisioner-terraform-state/$(TELEKUBE_CLUSTER_NAME)**
* AWS_KEY_NAME: SSH key for ec2 instance. The SSH key needs to be
  pre-created in AWS
* TELEKUBE_OPS_TOKEN: an agent token of telekube cluster
* TELEKUBE_NODE_PROFILE_COUNT_node: how many nodes
* TELEKUBE_NODE_PROFILE_INSTANCE_TYPE_node: node instance type

## Usage

Provisioner is deployed as a Docker image to a registry. We can
use Kubernetes job to execute Make target from the image.

An example job may look like this:

```
apiVersion: batch/v1
kind: Job
metadata:
  name: provision
  namespace: default
spec:
  activeDeadlineSeconds: 240
  template:
    spec:
      restartPolicy: OnFailure
      containers:
      - name: provision
        image: quay.io/gravitational/provisioner:0.0.3
        imagePullPolicy: Always
        args: ['cluster-provision']
        volumeMounts:
        - mountPath: /mnt/state
          name: state-volume
      volumes:
      - name: state-volume
        emptyDir: {}
```

The job can be run as an [application
hook](http://gravitational.com/docs/pack/#application-hooks)

### Tasks

Provisioner image has 4 main tasks:

* cluster-provision: Generate Terraform script, sync this script to S3, then
	combine this with other static Terraform templates and execute to form a cluster
  and its instance. Environment variables:

    * AWS_VPC_ID: existing VPC ID. If empty, a new VPC is created

* cluster-deprovision: Sync the generated Terraform before back from S3,
  then execute to destroy the whole cluster.
* nodes-provision: Do same thing as cluster-provision.
* nodes-deprovision: Remove an instance referenced in the environment
  variable `AWS_INSTANCE_PRIVATE_IP`.

## Customize Terraform script

We can customize Terraform script in `scripts/terraform/templates` which
are written as Go templates. Below is the list of variables that can be
used to override the templates:

* variables: a map with this only member

  * aws: a map with below member

    * subnets: an array of subnet CIDRs as strings
    * public_subnets: an array of string of public subnets CIDR
    * region: AWS region the VPC belongs to
    * vpc_id: VPC id, can be empty  if the new VPC is to be created
    * internet_gateway_id: internet gateway id, can be empty when not
      passing VPC ID
    * nat_gateways: an array of NAT gateway as strings on each of
      availabity zone, can be empty when not passing VPC ID
    * azs: an array of string of availability zones

If changes are made to the Terraform script, the docker image needs to
be rebuilt and published to the quay.io repository.

## Development

List of development tasks.

### Deploying to docker registry

```
make build-provisioner publish-provisioner
```

### Run Test

```
make test
```

### Manually testing

It is recommended to run manual tests during development. It is easier
if you collect the necessary environment variables into a file named
`dev.env`:

```
AWS_REGION=xxx
AWS_ACCESS_KEY_ID=xxx
AWS_SECRET_ACCESS_KEY=xxx
TELE_FLAGS=
OPS_URL=xxx
OPS_TOKEN=xxx
BUCKET_NAME=xxx
AWS_VPC_ID="xxxx"
TELEKUBE_CLUSTER_NAME=xxx
TELEKUBE_OPS_TOKEN=xxx
```

then pass it to docker:

```
docker run --rm -it \
  -v `pwd`/foo:/mnt/state/cluster \
  --env-file ./dev.env \
  quay.io/gravitational/provisioner:0.0.3 \
    init-cluster
```


### AWS Auto Scale Groups Support

Provisioner supports [AWS Auto Scaling Groups](http://docs.aws.amazon.com/autoscaling/latest/userguide/AutoScalingGroup.html).

Every Telekube cluster publishes join token and internal load balancer address via [Systems manager parameter store](http://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html).

Nodes provisioned via Auto Scaling group use this information to discover the cluster and join to it.

Provisioner scripts create IAM policies for nodes to read the SSM parameters and for master nodes to publish parameters to the store.

Master nodes are not part of the auto scaling group and are managed separately.



