# This Makefile is used by CI/CD builds.
#
# Prerequisites:
#     - Docker 1.9.1 or newer
#     - User must be a part of 'docker' group to use Docker without sudo
#

.DEFAULT_GOAL := ci

# URL of OPS server to build off of and upload app to
OPS_URL ?=
# OPS Center token to login
OPS_KEY ?=
# The tag on quay.io of the provisioning container to use
TAG ?=


TELE_FLAGS :=
CWD := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

TELE ?= $(shell command -v tele;)
ifndef TELE
$(error "tele is not available, please install using this docs http://gravitational.com/docs/quickstart/#getting-the-tools")
endif


.PHONY: ci
ci: tele-login
	$(eval TMP := $(shell mktemp -d))
	mkdir -p $(TMP)/resources/
	cp $(CWD)/resources/* $(TMP)/resources/
	find $(TMP)/resources/ -type f -name '*.yaml' -exec sed -i 's|XXX_PROVISIONER_VERSION_XXX|$(TAG)|' {} \;
	$(TELE) $(TELE_FLAGS) build \
        	--skip-version-check \
		--output="$(TMP)/installer.tar" \
		--version="1.0.0-$(TAG)" \
		--overwrite \
		--repository="https://$(OPS_URL)" \
		"$(TMP)/resources/app.yaml"
	$(TELE) $(TELE_FLAGS) push --force "$(TMP)/installer.tar"
	rm -r $(TMP)


.PHONY: tele-login
tele-login: check-env
	$(TELE) login --ops=${OPS_URL} --key=${OPS_KEY}

.PHONY: check-env
check-env:
ifndef OPS_URL
	$(error OPS_URL is undefined)
endif
ifndef OPS_KEY
	$(error OPS_KEY is undefined)
endif
ifndef TAG
	$(error TAG is undefined)
endif
