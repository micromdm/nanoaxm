# NanoAXM

[![CI/CD](https://github.com/micromdm/nanoaxm/actions/workflows/on-push-pr.yml/badge.svg)](https://github.com/micromdm/nanoaxm/actions/workflows/on-push-pr.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/micromdm/nanoaxm.svg)](https://pkg.go.dev/github.com/micromdm/nanoaxm)

NanoAXM is a set of tools and a Go libraries powering them for communicating with [Apple School and Business Manager API](https://developer.apple.com/documentation/apple-school-and-business-manager-api).

"AxM" is a sort of initialism representing both "Apple Business Manager" and "Apple School Manager" together.

## Getting started & Documentation

- [Quickstart](docs/quickstart.md)  
A guide to get NanoAXM up and running quickly.

- [Operations Guide](docs/operations-guide.md)  
A brief overview of the various tools and utilities for working with NanoAXM.

## Getting the latest version

* Release `.zip` files containing the project should be attached to every [GitHub release](https://github.com/micromdm/nanoaxm/releases).
  * Release zips are also [published](https://github.com/micromdm/nanoaxm/actions) for every `main` branch commit.
* A Docker container is built and [published to the GHCR.io registry](http://ghcr.io/micromdm/nanoaxm) for every release.
  * `docker pull ghcr.io/micromdm/nanoaxm:latest` — `docker run ghcr.io/micromdm/nanoaxm:latest`
  * A Docker container is also published for every `main` branch commit (and tagged with `:main`)
* If you have a [Go toolchain installed](https://go.dev/doc/install) you can checkout the source and simply run `make`.

## Tools and utilities

NanoAXM contains a few tools and utilities. At a high level:

- **API configuration & reverse proxy server.** The primary Go server component is used for configuring NanoAXM and talking with Apple's AxM servers. It hosts its own API for configuring OAuth 2 crendetials (called "AxM names") and also hosts a transparently authenticating reverse proxy for talking 'directly' to Apple's ABM and ASM (AxM) API endpoints.
- **Scripts, tools, and helpers.** A set of [tools](tools) and utilities for talking to the Apple AxM API services — mostly implemented as shell scripts that communicate with the Go server.

See the [Operations Guide](docs/operations-guide.md) for more details and usage documentation.

## Go library

NanoAXM is also a Go library for accessing the Apple AxM APIs. There are two components to the Go library:

* The higher-level [goaxm](https://pkg.go.dev/github.com/micromdm/nanoaxm/goaxm) package implements Go methods and structures for talking to the individual AxM API endpoints.
* The lower-level [client](https://pkg.go.dev/github.com/micromdm/nanoaxm/client) package implements primitives, helpers, and middleware for authenticating to the AxM APIs and managing OAuth 2 tokens.

See the [Go Reference documentation](https://pkg.go.dev/github.com/micromdm/nanoaxm) (or the Go source itself, of course) for details on these packages.
