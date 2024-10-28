package main

import (
    "os"
    "path/filepath"
    "errors"
    "encoding/json"
    "fmt"
)

type latestPayload struct {
    Version string `json:"version"`
}

func readLatestFile(lat_path string) (latestPayload, error) {
    output := latestPayload{}
    if _, err := os.Stat(lat_path); errors.Is(err, os.ErrNotExist) {
        return output, nil
    }

    handle, err := os.Open(lat_path)
    if err != nil {
        return output, fmt.Errorf("failed to open %q; %w", lat_path, err)
    }
    defer handle.Close()

    dec := json.NewDecoder(handle)
    err = dec.Decode(&output)
    if err != nil {
        return output, fmt.Errorf("failed to parse %q; %w", lat_path, err)
    }

    return output, nil
}

func ignoreNonLatest(registry, project, asset string) (bool, error) {
    asset_dir := filepath.Join(registry, project, asset)

    lat_path := filepath.Join(asset_dir, "..latest")
    payload, err := readLatestFile(lat_path)
    if err != nil {
        return false, err
    }

    all_errors := []error{}
    known_versions, err := os.ReadDir(asset_dir)
    if err != nil {
        return false, fmt.Errorf("failed to read versions of %q; %w", asset_dir, err)
    }

    changed := false
    for _, x := range known_versions {
        if !x.IsDir() {
            continue
        }

        v := x.Name()
        ipath := filepath.Join(asset_dir, v, ".SewerRatignore")
        _, err := os.Stat(ipath)

        if v == payload.Version {
            if !errors.Is(err, os.ErrNotExist) {
                changed = true
                err := os.Remove(ipath)
                if err != nil {
                    all_errors = append(all_errors, fmt.Errorf("failed to remove ignore file at %q; %w", ipath, err))
                }
            }
        } else {
            if errors.Is(err, os.ErrNotExist) {
                changed = true
                err := os.WriteFile(ipath, []byte{}, 0644)
                if err != nil {
                    all_errors = append(all_errors, fmt.Errorf("failed to write ignore file to %q; %w", ipath, err))
                }
            }
        }
    }

    if len(all_errors) > 0 {
        return changed, errors.Join(all_errors...)
    } else {
        return changed, nil
    }
}
