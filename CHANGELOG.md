# Changelog

## Unreleased

### Fixed

- The exporter could crash when a Redfish collection contained links to unexpected paths.
  Since the `host` parameter was not validated, this could be triggered by anyone able to reach the exporter.
- The `host` parameter is now validated to be `host[:port]`. Previously a crafted parameter could send the
  configured credentials to an arbitrary host or path. Requests without or with an invalid `host` are answered
  with `400 Bad Request` (was `500`) and are no longer logged as errors.
- `ilo_storage_disk_capacity_byte` was 1024 times too small for drives reported via the HPE SmartStorage API (iLO 4/5).
- `SIGINT` (Ctrl-C) did not stop the exporter, `SIGTERM` did not shut it down gracefully and pending traces were not flushed.
- `API_USERNAME`/`API_PASSWORD` overwrote both credentials when only one of `-api.username`/`-api.password` was set.
- The version reported by release binaries and container images is now set at build time.

### Added

- `-web.allowed-targets` restricts the hosts which can be scraped to a list of host names, IP addresses and CIDR prefixes.

### Changed

- The binary in release archives is named `ilo_exporter` (was `ilo5_exporter`), matching the container image.
