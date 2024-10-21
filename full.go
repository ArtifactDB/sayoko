package main

import (
    "os"
    "path/filepath"
    "errors"
    "fmt"
)

func fullScan(registry string, url string, to_reignore map[string]bool) error {
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
        is_reg, err := isProjectRegistered(registry, url, project)
        if err != nil {
            all_errors = append(all_errors, err)
            continue
        }
        if !is_reg {
            continue
        }

        proj_path := filepath.Join(registry, project)
        asses, err := os.ReadDir(proj_path)
        if err != nil {
            all_errors = append(all_errors, fmt.Errorf("failed to read contents of %q; %w", proj_path, err))
            continue
        }

        for _, ass := range asses {
            if !ass.IsDir() {
                continue
            }
            to_reignore[project + "/" + ass.Name()] = true
        }
    }

    if len(all_errors) > 0 {
        return errors.Join(all_errors...)
    } else {
        return nil
    }
}
