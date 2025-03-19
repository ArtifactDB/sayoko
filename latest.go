package main

import (
    "os"
    "path/filepath"
    "errors"
    "encoding/json"
    "fmt"
)

type latestInfo struct {
    Version string `json:"version"`
}

func readLatestFile(lat_path string) (latestInfo, error) {
    output := latestInfo{}
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

func ignoreNonLatest(rest_url, asset_dir string) error {
    lat_path := filepath.Join(asset_dir, "..latest")
    payload, err := readLatestFile(lat_path)
    if err != nil {
        return err
    }

    registered_versions, err := listRegisteredSubdirectories(rest_url, asset_dir)
    if err != nil {
        return fmt.Errorf("failed to list registered versions of %q; %w", asset_dir, err)
    }

    all_errors := []error{}
    for _, ver := range registered_versions {
        if ver == payload.Version {
            continue
        }
        version_dir := filepath.Join(asset_dir, ver)
        regerr := registerDirectory(rest_url, version_dir, false)
        if regerr != nil {
            all_errors = append(all_errors, regerr)
        }
    }

    if payload.Version != "" {
        version_dir := filepath.Join(asset_dir, payload.Version)
        regerr := registerDirectory(rest_url, version_dir, true)
        if regerr != nil {
            all_errors = append(all_errors, regerr)
        }
    }

    if len(all_errors) > 0 {
        return errors.Join(all_errors...)
    } else {
        return nil
    }
}
