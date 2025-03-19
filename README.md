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

By default, this will scan the log directory every 10 minutes and will do a full registry check every 24 hours.
These intervals can be modified with the `-log` and `-full` flags, respectively.

After every log scan, **sayoko** produces a `.sayoko_last_scan` file containing the RFC3339-formatted time of the most recent log.
This avoids redundant re-processing of the same log files when **sayoko** itself is restarted.
Advanced users can exploit this by modifying the timestamp in this file to force **sayoko** to process logs after a desired timepoint.
The path of this timestamp file can also be modified via the `-timestamp` flag.

## Developer notes

Download the latest [SewerRat binary](https://github.com/ArtifactDB/SewerRat/releases/tag/latest) and run it with default arguments.
Once the SewerRat service has started successfully, testing can be performed with the usual `go test` commands.
