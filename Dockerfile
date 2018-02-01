FROM quay.io/gravitational/debian-grande:0.0.1

# What Terraform version to install
ARG TERRAFORM_VER

RUN ( \
        apt-get update && \
        apt-get install --yes --no-install-recommends \
            make curl unzip python python-pip curl unzip groff\
    )

RUN ( cd /usr/local/bin && \
     curl -O https://releases.hashicorp.com/terraform/${TERRAFORM_VER}/terraform_${TERRAFORM_VER}_linux_amd64.zip && \
     unzip terraform_${TERRAFORM_VER}_linux_amd64.zip && \
     rm -f terraform_${TERRAFORM_VER}_linux_amd64.zip )

RUN curl "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip" -o "awscli-bundle.zip"&& \
   unzip awscli-bundle.zip && \
   ./awscli-bundle/install -i /usr/local/aws -b /usr/local/bin/aws

RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# bundle Makefile and Terraform
ADD scripts /usr/local/bin/provisioner

# bundle the main provisioner program
ADD build/provisioner /usr/local/bin/inspect

# By setting this entry point, we expose make target as command
ENTRYPOINT ["/usr/bin/dumb-init", "/usr/bin/make", "-C", "/usr/local/bin/provisioner"]
