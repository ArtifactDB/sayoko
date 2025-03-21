package main

import (
    "os"
    "path/filepath"
    "time"
    "encoding/json"
    "fmt"
    "strings"
    "errors"
)

type logEntry struct {
    Type string `json:"type"`
    Project string `json:"project"`
    Asset string `json:"asset"`
}

func readLog(logpath string) (logEntry, error) {
    output := logEntry{}

    handle, err := os.Open(logpath)
    if err != nil {
        return output, fmt.Errorf("failed to open %q; %w", logpath, err)
    }
    defer handle.Close()

    dec := json.NewDecoder(handle)
    err = dec.Decode(&output)
    if err != nil {
        return output, fmt.Errorf("failed to parse %q; %w", logpath, err)
    }

    return output, nil
}

func processLogs(rest_url string, registry string, names []string, last_scan time.Time) (time.Time, error) {
    lpath := filepath.Join(registry, "..logs")
    dirhandle, err := os.Open(lpath)
    if err != nil {
        return last_scan, fmt.Errorf("failed to open directory handle for %q; %w", lpath, err)
    }
    defer dirhandle.Close()

    lognames, err := dirhandle.Readdirnames(0)
    if err != nil {
        return last_scan, fmt.Errorf("failed to read log directory at %q; %w", lpath, err)
    }

    all_errors := []error{}
    latest := last_scan

    // No need to process things in order as long as we get everything past the last scan's timepoint;
    // all directories will converge to being registered or not, so it doesn't matter.
    for _, n := range lognames {
        pos := strings.IndexByte(n, '_')
        if pos < 0 {
            all_errors = append(all_errors, fmt.Errorf("failed to parse time for %q; %w", n, err))
            continue 
        }
        stamp_str := n[:pos]
        stamp, err := time.Parse(time.RFC3339, stamp_str)
        if err != nil {
            all_errors = append(all_errors, fmt.Errorf("failed to parse time for %q; %w", n, err))
            continue
        }
        if !stamp.After(last_scan) {
            continue
        }
        if stamp.After(latest) {
            latest = stamp
        }

        logpath := filepath.Join(lpath, n)
        payload, err := readLog(logpath)
        if err != nil {
            all_errors = append(all_errors, err)
            continue
        }

        if payload.Type == "add-version" || payload.Type == "delete-version" || payload.Type == "reindex-version" {
            if payload.Project == "" || payload.Asset == "" {
                all_errors = append(all_errors, fmt.Errorf("empty project/asset fields in %q", logpath))
                continue
            }
            err := ignoreNonLatest(
                rest_url,
                filepath.Join(registry, payload.Project, payload.Asset),
                names,
                (payload.Type == "reindex-version"), // Immediately pick up any changes from reindexing.
            )
            all_errors = append(all_errors, err)

        } else if payload.Type == "delete-asset" {
            if payload.Project == "" || payload.Asset == "" {
                all_errors = append(all_errors, fmt.Errorf("empty project/asset fields in %q", logpath))
                continue
            }
            err := deregisterAllSubdirectories(rest_url, filepath.Join(registry, payload.Project, payload.Asset))
            all_errors = append(all_errors, err)

        } else if payload.Type == "delete-project" {
            if payload.Project == "" {
                all_errors = append(all_errors, fmt.Errorf("empty project field in %q", logpath))
                continue
            }
            err := deregisterAllSubdirectories(rest_url, filepath.Join(registry, payload.Project))
            all_errors = append(all_errors, err)
        }
    }

    if len(all_errors) > 0 {
        return latest, errors.Join(all_errors...)
    } else {
        return latest, nil
    }
}
