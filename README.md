# Paperless-ngx metrics for Prometheus

[![Latest release](https://img.shields.io/github/v/release/hansmi/prometheus-paperless-exporter)][releases]
[![Release workflow](https://github.com/hansmi/prometheus-paperless-exporter/actions/workflows/release.yaml/badge.svg)](https://github.com/hansmi/prometheus-paperless-exporter/actions/workflows/release.yaml)
[![CI workflow](https://github.com/hansmi/prometheus-paperless-exporter/actions/workflows/ci.yaml/badge.svg)](https://github.com/hansmi/prometheus-paperless-exporter/actions/workflows/ci.yaml)
[![Go reference](https://pkg.go.dev/badge/github.com/hansmi/prometheus-paperless-exporter.svg)](https://pkg.go.dev/github.com/hansmi/prometheus-paperless-exporter)

This repository hosts a Prometheus metrics exporter for
[Paperless-ngx][paperless], a document management system transforming physical
documents into a searchable online archive. The exporter relies on [Paperless'
REST API][paperless-api].

An implementation using the API was chosen to provide the same perspective as
web browsers.


## Usage

`prometheus-paperless-exporter` listens on TCP port 8081 by default. To listen on
another address use the `-web.listen-address` flag (e.g.
`-web.listen-address=127.0.0.1:3000`).

TLS and HTTP basic authentication is supported through the [Prometheus exporter
toolkit][toolkit]. A configuration file can be passed to the `-web.config` flag
([documentation][toolkitconfig]).

See the `--help` output for more flags.

### Remote version polling

When `--enable-remote-network` is enabled the exporter will periodically poll
the Paperless API to check whether a newer Paperless-ngx version is available.
The polling interval duration can be configured with `--remote-version-interval`.
Default is `1h`.

## Collector selection

The exporter supports configuring which collectors are enabled via the 
`--collectors` flag. Provide a comma-separated list of collector ids.
If the flag is omitted or empty, all standard collectors are enabled.

Available collector ids:

* `tag`
* `correspondent`
* `document_type`
* `storage_path`
* `task`
* `log`
* `group`
* `user`
* `document`
* `status`
* `statistics`
* `remote_version` (requires `--enable-remote-network` to actually be used)

Examples:

Enable only tags and documents:

```shell
./prometheus-paperless-exporter --collectors=tag,document
```

Enable all standard collectors and also the remote version collector
(remote network checks must be explicitly allowed):

```shell
./prometheus-paperless-exporter --enable-remote-network
```

If you specify unknown collector ids the exporter will exit with an error 
listing the unknown and known ids.

## Permissions

The metrics user requires [_view_ permissions][paperless-permissions] on the
following object types:

* Admin
  * Required for log analysis.
  * Starting with version 2.8 there is no distinction between different access
    modes ([paperless-ngx#6380](https://github.com/paperless-ngx/paperless-ngx/pull/6380)).
* Correspondent
* Document
* DocumentType
* Group
* PaperlessTask
* StoragePath
* Tag
* User


## Installation

Pre-built binaries are provided for [all releases][releases]:

* Binary archives (`.tar.gz`)
* Debian/Ubuntu (`.deb`)
* RHEL/Fedora (`.rpm`)
* Microsoft Windows (`.zip`)

Docker images via GitHub's container registry. The image supports Linux/AMD64
and Linux/ARM64.

```shell
docker pull ghcr.io/hansmi/prometheus-paperless-exporter
```

It's also possible to produce custom builds directly using [Go][golang] or
[GoReleaser][goreleaser].


### Docker Compose

An example configuration for [Docker Compose][dockercompose] is available in
`contrib/docker-compose`:

```shell
env --chdir contrib/docker-compose docker-compose up
```


[dockercompose]: https://docs.docker.com/compose/
[golang]: https://golang.org/
[goreleaser]: https://goreleaser.com/
[paperless-api]: https://docs.paperless-ngx.com/api/
[paperless]: https://docs.paperless-ngx.com/
[paperless-permissions]: https://docs.paperless-ngx.com/usage/#permissions
[releases]: https://github.com/hansmi/prometheus-paperless-exporter/releases/latest
[toolkit]: https://github.com/prometheus/exporter-toolkit
[toolkitconfig]: https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md

<!-- vim: set sw=2 sts=2 et : -->
