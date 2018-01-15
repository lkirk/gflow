### -=<(gflow)>=-
DEFAULT_GOAL: go-build
.PHONY: go-build clean

WD:=$(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
SHELL:=/bin/bash -eo pipefail
SUB_MAKE_OPTS:=--no-print-directory
MAKE+=$(SUB_MAKE_OPTS)

gflow:=$(WD)/gflow

go-build: $(gflow)

.PHONY: test
TEST_OPTS:=
test:
	go test ./... $(TEST_OPTS)

.PHONY: test
LINT_OPTS:=
lint:
	golint $(LINT_OPTS)

dev:
	$(MAKE) test lint TEST_OPTS+='-v -race'
	go vet

.PHONY: coverage-report
COVERAGE_REPORT_OPTS:=
_COVERAGE_REPORT:=coverage.out
coverage-report:
	$(MAKE) test TEST_OPTS+=-coverprofile=$(_COVERAGE_REPORT)
	go tool cover -html=$(_COVERAGE_REPORT)

## always build on make
.PHONY: $(gflow)
$(gflow):
	CGO_ENABLED=0 go build

## remove binary
.PHONY: clean
clean:
	rm -f $(_COVERAGE_REPORT)
	rm -f $(gflow)
	rm -rf $(WD)/testoutput

## release
RELEASE-INCREMENTS:=major minor patch

.PHONY: release-
release-:
	$(info === release a new version ===)
	$(info Use one of the 3 options :'$(RELEASE-INCREMENTS)' to formulate the make command.)
	$(info For example to make a '$(lastword $(RELEASE-INCREMENTS))' release, run: 'make release-$(lastword $(RELEASE-INCREMENTS))')
	@echo '' > /dev/null ## suppress Nothing to be done for 'release-'. message

define release_template =
release-$(1):
	@ \
	set -x \
	git checkout dev ;\
	git pull ;\
	git checkout master ;\
	git pull ;\
	git pull --tags ;\
	NEW_VERSION=$$$$(git describe | ./scripts/increment-version $(1)) ;\
	git checkout dev ;\
	sed -i -re"s/[0-9]+\.[0-9]+\.[0-9]+/$$$$NEW_VERSION/g" \
		$(WD)/README.org ;\
	git push ;\
	git checkout master ;\
	git merge --no-ff -m'Merge dev into master by Makefile' dev ;\
	git tag -a -m'Increment $(1) version by Makefile' $$$$NEW_VERSION ;\
	git push --tags ;\
	git push ;\
	git checkout dev
endef

$(foreach increment,$(RELEASE-INCREMENTS),$(eval $(call release_template,$(increment))))
