# OCI image registry compatibility tests

A go package that tests the conformance of a container registry with the OCI image specification.

## Prerequesites

You will need a working go install. See: https://go.dev/doc/install

After that you can install the dependencies by changing into this project and running
```bash
go mod download
```

## Run tests

There are four environment variables used to configure the registry under test:
```bash
export REGISTRY_HOST="host"
export REGISTRY_USER="user@mail.org"
export REGISTRY_PASSWORD="your-access-token"
export REGISTRY_NAMESPACE="namespace/project"
```

Run the tests with:
```bash
go test -v ./...
```

## License

This project is licensed according to the terms of the Apache License, v. 2.0.
A copy of the license is provided in the `LICENSE` file.
