# Gobbler-SewerRat integration

[![Test and build](https://github.com/ArtifactDB/sayoko/actions/workflows/build.yaml/badge.svg)](https://github.com/ArtifactDB/sayoko/actions/workflows/build.yaml)
[![Publish version](https://github.com/ArtifactDB/sayoko/actions/workflows/publish.yaml/badge.svg)](https://github.com/ArtifactDB/sayoko/actions/workflows/publish.yaml)
[![Latest version](https://img.shields.io/github/v/tag/ArtifactDB/sayoko?label=Version)](https://github.com/ArtifactDB/sayoko/releases)

## Overview

**sayoko** ensures that only the latest version of each [Gobbler](https://github.com/ArtifactDB/gobbler) asset is included in the [SewerRat](https://github.com/ArtifactDB/SewerRat) index.
It does so by registering the subdirectory corresponding to the latest version of each asset and deregistering everything else.
The aim is to provide users with a more up-to-date search of Gobbler assets via SewerRat.
**sayoko** tracks changes in the Gobbler registry by scanning the log directory for updates.
It will also periodically check the entire Gobbler registry to ensure that the latest version is correctly registered.

## Instructions

The usual `go build .` command produces the `sayoko` binary.
We can then run it as shown below, using an account that has write permissions to the Gobbler registry.

```bash
./sayoko \
    -registry PATH_TO_GOBBLER_REGISTRY
    -url URL_FOR_SEWERRAT_REST_API
```

Options include:

- `-names`, a comma-separated list of names of metadata files to be indexed.
  If not provided, this defaults to `metadata.json`.
- `-log`, the interval between scans of the Gobbler log directory, in minutes.
  This defaults to 10 minutes.
- `-full`, the interval between full scans of the Gobbler registry, in hours.
  This defaults to 168 hours (i.e., weekly).
- `-timestamp`, a path to a file in which **sayoko** can store the timestamp of the last log scan.
  This defaults to `.sayoko_last_scan`.

More specifically: after every log scan, **sayoko** produces a timestamp file containing the RFC3339-formatted time of the most recent log.
This prevents redundant re-processing of the same log files when **sayoko** itself is restarted.
Advanced users can exploit this by modifying the timestamp in this file to force **sayoko** to process logs after a desired timepoint.

## Developer notes

Download the latest [SewerRat binary](https://github.com/ArtifactDB/SewerRat/releases/tag/latest) and run it with default arguments.
Once the SewerRat service has started successfully, testing can be performed with the usual `go test` commands.
