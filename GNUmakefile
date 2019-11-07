TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PROJ_NAME=terraform-provider-okta
TESTVARS=TF_ACC=OKTA_API_KEY=somethingelse

default: build
.PHONY: build build/nix build/win test testacc vet fmt fmtcheck errcheck

build: build/nix build/win

build/nix: fmtcheck
	GOOS=linux go build

build/win: fmtcheck
	GOOS=windows go build

get:
	go mod download

docker:
	docker run --rm -it \
		-v $(PWD):/srv/Brightspace/$(PROJ_NAME) \
		--workdir /srv/Brightspace/$(PROJ_NAME) \
		golang bash

test: fmtcheck
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: fmtcheck
	$(TESTVARS) go test $(TEST) -v $(TESTARGS) -timeout 120m

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"
