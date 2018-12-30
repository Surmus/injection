GO ?= go
GOFMT ?= gofmt "-s"
VETPACKAGES ?= $(shell $(GO) list ./... | grep -v /vendor/)
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")
TESTFOLDER := $(shell $(GO) list ./... | grep -v test)

all: install

install: deps
	dep ensure

.PHONY: test
test:
	echo "mode: count" > coverage.out
	for d in $(TESTFOLDER); do \
		$(GO) test -v -covermode=count -coverprofile=profile.out $$d > tmp.out; \
		cat tmp.out; \
		if grep -q "^--- FAIL" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		elif grep -q "build failed" tmp.out; then \
			rm tmp.out; \
			exit; \
		fi; \
		if [ -f profile.out ]; then \
			cat profile.out | grep -v "mode:" >> coverage.out; \
			rm profile.out; \
		fi; \
	done

.PHONY: fmt
fmt:
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

vet:
	$(GO) vet $(VETPACKAGES)

deps:
	@hash dep > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh; \
	fi