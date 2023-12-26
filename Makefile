.PHONY: fmt
fmt:
	golangci-lint run --fix --config ./.golangci.yml

.PHONY: test
test:
	$(MAKE) -C testdata vendor
	go test -v ./...

.PHONY: install
install:
	go install ./cmd/spancheck
	@echo "Installed in $(shell which spancheck)"