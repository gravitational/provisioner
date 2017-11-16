# This Makefile is used by CI/CD builds.
#
# Prerequisites:
#     - Docker 1.9.1 or newer
#     - You must be a part of 'docker' group to use Docker without sudo
#

.DEFAULT_GOAL := ci

OPS_URL :=
OPS_KEY :=
TAG :=
TELE_FLAGS :=
CWD := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
APP_VERSION ?= $(shell git -C $(CWD) describe --tags)


.PHONY: ci
ci: tele-login
	$(eval TMP := $(shell mktemp -d))
	mkdir -p $(TMP)/resources/
	cp $(CWD)/resources/* $(TMP)/resources/
	find $(TMP)/resources/ -type f -name '*.yaml' -exec sed -i 's|XXX_PROVISIONER_VERSION_XXX|$(TAG)|' {} \;
	tele $(TELE_FLAGS) build \
        --skip-version-check \
		--output="$(TMP)/installer.tar" \
		--version="$(APP_VERSION)" \
		--overwrite "$(TMP)/resources/app.yaml" --repository="https://$(OPS_URL)"
	tele $(TELE_FLAGS) push --force "$(TMP)/installer.tar"
	rm -r $(TMP)


.PHONY: tele-login
tele-login: check-env
	tele login --ops=${OPS_URL} --key=${OPS_KEY}

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
