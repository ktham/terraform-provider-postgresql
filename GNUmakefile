default: build

build:
	go build -v ./...

# Compiles the provider and places the binary in the directory specified by the GOBIN env var ($HOME/go/bin by default)
install: build
	go install -v ./...

# Generate docs and places them in the 'docs/' directory
generate:
	cd tools; go generate ./...

# Run acceptance tests
testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: \
	build \
	generate \
	install \
	testacc \
	testacc_crdb \
