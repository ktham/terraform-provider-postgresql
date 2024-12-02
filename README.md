# Terraform Provider for PostgreSQL

This Terraform provider is used to manage PostgreSQL objects.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.11
- [Go](https://golang.org/doc/install) >= 1.23
- [podman](https://golang.org/doc/install) >= 5.25
    * On MacOS: Install from https://podman.io/
    * On Ubuntu: `apt-get install -y podman`
- [podman-compose](https://golang.org/doc/install) >= 1.2.0
    * On MacOS: `brew install podman-compose`
    * On Ubuntu: `apt-get install -y python3-pip` then `pip3 install podman-compose`

## Development
### Building The Provider

You can build the provider by running `make build`. It should build cleanly without errors.

### Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

### Building a local development version of the provider
If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `make install`.
This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To use your locally built provider, you'll need to specify an override in your `~/.terraformrc` file:

```
provider_installation {
  dev_overrides {
     // Note: Replace `${GO_BIN_PATH}` with the fully qualified path on your system to your Go Bin directory
     // For example: "
    "ktham/postgresql" = "${GO_BIN_PATH}"
  }

  direct {}
}
```

### Starting up a local Postgres database server
You will also need to start a PostgreSQL container that can then be used for development and testing.

```shell 
podman compose up -d
```

### Testing
To run the full suite of Acceptance tests, make sure you have a working Postgres server running, then run `make testacc`.

## Generating documentation
This provider uses terraform-plugin-docs to generate documentation and store it in the docs/ directory.

Use `make generate` to ensure the documentation is regenerated with any changes.
