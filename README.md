# provisioner
Terraform based provisioners for ops center

# Overview

Provisioner include two components:

- an executable binary to generate terraform script
- a docker image bundles above binary and makefile expose set of tasks
  to provision cluster

The docker image is the final package to be deploy and executed with
kubernetes job to create VPC. The docker entrypoint is set to **Make**
so you can invoke any make target by passing command when executing with
docker run.

Provisioner works by generating necessary Terraform script and apply
those Terraform script. The script also sync to a S3 bucket. And re-sync
from S3 to local whenever it is called again, such as when destroying
cluster.

Provisioner can run standalone, however, its main usage is to be
integrate with Telekube application hook to create/reuse VPC and launch
instance into the VPC.

Provisioner works in two modes:

1. Using an existing VPC
2. Creating a new VPC

## Existing VPC

In this mode, provisioner reads VPC configuration and reuse NAT gateway,
internet gateway. The generated Terraform will contain new resources for
subnets, security groups. The amount of pair of private/public subnet is
equal the amout of NAT gateways of the existing VPC. The subnet CIDR is
calculated and will not of couse not overlap with any existing subnets in VPC.

## Creating a new VPC

In this mode, provisioner simply generates TerraForm script with all new
resource to create a VPC:

  - NAT gateway and its Elastic IP respectively
  - Internet gateway and its Elastic IP respectively
  - Subnets:Private and Public and their route tables respectively

Provisioner creates subnet on all availability zone of region, and runs
a NAT gateway on each availability zone.

## Environment variable

Please refer to `scripts/Makefile` for a list of available environment variable.
Below are required variable for all tasks:

* AWS_REGION: aws region of S3 bucket to store Terraform script and
  state.
* AWS_ACCESS_KEY_ID: aws access key id
* AWS_SECRET_ACCESS_KEY: aws secret key
* TELEKUBE_CLUSTER_NAME: telekube cluster. We will store generated
  script into an bucket in form of **terraform-cluster-state-$(TELEKUBE_CLUSTER_NAME)**
* AWS_KEY_NAME: SSH key for ec2 instance. The SSH Key needs to be
  pre-created in AWS
* TELEKUBE_OPS_TOKEN: an agent token of telekube cluster
* TELEKUBE_NODE_PROFILE_COUNT_node: how many node
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
	combine this with other static Terraform template and execute to form a cluster
  and its instance. Environment variables:

    * AWS_VPC_ID: existing VPC ID, if empty will create a new vpc

* cluster-deprovision: Sync the generated Terraform before back from S3,
  then execute to destroy the whole cluster.
* nodes-provision: Do same thing as cluster-provision.
* nodes-deprovision: Remove an instance which is referenced from
  environment variable `AWS_INSTANCE_PRIVATE_IP`.

## Customize Terraform script

We can customize Terraform script in `scripts/terraform/templates`. It's
basically a Golang template for Terraform. Below are list of variable we
can use:

* variables: a map with this only member

  * aws: a map with below member

    * subnets: an array of string of subnets CIDR
    * public_subnets: an array of string of public subnets CIDR
    * region: aws region the VPC belongs to
    * vpc_id: vpc id, can be empty depend on what we pass
    * internet_gateway_id: internet gateway id, can be empty when not
      passing VPC ID
    * nat_gateways: an array of string of NAT gateway on each of
      availabity zone, can be empty when not passing VPC ID
    * azs: an array of string of availability zone

Once we make change to Terraform, we also need to rebuild the image and
publish it to Quay registry.

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

It's useful to invoke manually testing during development. We can easily
do that by create a list of enviroment in a file says `dev.env`:

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

and manually invoke them with docker:

```
docker run --rm -it \
  -v `pwd`/foo:/mnt/state/cluster \
  --env-file ./dev.env \
  quay.io/gravitational/provisioner:0.0.3 \
    init-cluster
```
