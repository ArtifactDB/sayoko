package main

import (
    "os"
    "path/filepath"
    "errors"
    "fmt"
)

func fullScan(rest_url string, registry string) error {
    contents, err := os.ReadDir(registry) 
    if err != nil {
        return fmt.Errorf("failed to read the registry contents; %w", err)
    }

    all_errors := []error{}
    for _, proj := range contents {
        if !proj.IsDir() {
            continue
        }
        project := proj.Name()
        project_dir := filepath.Join(registry, project)
        asses, err := os.ReadDir(project_dir)
        if err != nil {
            all_errors = append(all_errors, fmt.Errorf("failed to list assets for project %q; %w", project, err))
            continue
        }

        for _, ass := range asses {
            if !ass.IsDir() {
                continue
            }
            asset := ass.Name()
            asset_dir := filepath.Join(project_dir, asset)
            err := ignoreNonLatest(rest_url, asset_dir)
            all_errors = append(all_errors, err)
        }
    }

    // Put this _after_ we check that we can list the contents of the registry,
    // to avoid premature deregistration upon sporadic unmounting of the registry's FS.
    err = deregisterMissingSubdirectories(rest_url, registry)
    if err != nil {
        all_errors = append(all_errors, err)
    }

    if len(all_errors) > 0 {
        return errors.Join(all_errors...)
    } else {
        return nil
    }
}
