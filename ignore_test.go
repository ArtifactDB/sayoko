package main

import (
    "os"
    "path/filepath"
    "testing"
    "errors"
)

func TestReadLatestFile(t *testing.T) {
    workdir, err := os.MkdirTemp("", "")
    if err != nil {
        t.Fatalf("failed to create working directory; %v", err)
    }

    lat_path := filepath.Join(workdir, "..latest")
    latest, err := readLatestFile(lat_path)
    if err != nil {
        t.Fatalf("failed to read the ..latest file; %v", err)
    }
    if latest.Latest != "" {
        t.Fatalf("..latest file should be absent; %v", err)
    }

    err = os.WriteFile(lat_path, []byte("{ \"latest\": \"foobar\" }"), 0644)
    if err != nil {
        t.Fatalf("failed to write to the ..latest file; %v", err)
    }

    latest, err = readLatestFile(lat_path)
    if err != nil {
        t.Fatalf("failed to read the ..latest file; %v", err)
    }
    if latest.Latest != "foobar" {
        t.Fatalf("unexpected latest version %q", latest)
    }
}

func TestIgnoreNonLatest(t *testing.T) {
    registry, err := os.MkdirTemp("", "")
    if err != nil {
        t.Fatalf("failed to create registry; %v", err)
    }

    project := "leilla"
    project_dir := filepath.Join(registry, project)
    err = os.Mkdir(project_dir, 0755)
    if err != nil {
        t.Fatalf("failed to create project directory; %v", err)
    }

    asset := "kanon"
    asset_dir := filepath.Join(project_dir, asset)
    err = os.Mkdir(asset_dir, 0755)
    if err != nil {
        t.Fatalf("failed to create asset directory; %v", err)
    }

    lat_path := filepath.Join(asset_dir, "..latest")
    err = os.WriteFile(lat_path, []byte("{ \"latest\": \"3\" }"), 0644)
    if err != nil {
        t.Fatalf("failed to write to the ..latest file; %v", err)
    }

    for _, version := range []string{ "1", "2", "3" } {
        version_dir := filepath.Join(asset_dir, version)
        err = os.Mkdir(version_dir, 0755)
        if err != nil {
            t.Fatalf("failed to create asset directory; %v", err)
        }
    }

    t.Run("basic", func(t *testing.T) {
        changed, err := ignoreNonLatest(registry, project, asset)
        if err != nil {
            t.Fatalf("unexpected errors when ignoring non-latest version; %v", err)
        }
        if !changed {
            t.Fatal("expected some kind of change when ignoring non-latest version")
        }

        if _, err := os.Stat(filepath.Join(asset_dir, "1", ".SewerRatignore")); err != nil {
            t.Fatalf("expected a .SewerRatignore file for version 1")
        }
        if _, err := os.Stat(filepath.Join(asset_dir, "2", ".SewerRatignore")); err != nil {
            t.Fatalf("expected a .SewerRatignore file for version 2")
        }
        if _, err := os.Stat(filepath.Join(asset_dir, "3", ".SewerRatignore")); !errors.Is(err, os.ErrNotExist) {
            t.Fatalf("expected no .SewerRatignore file for version 3")
        }

        // Next run has no change.
        changed, err = ignoreNonLatest(registry, project, asset)
        if err != nil {
            t.Fatalf("unexpected errors when ignoring non-latest version; %v", err)
        }
        if changed {
            t.Fatal("expected no change when re-ignoring non-latest version")
        }
    })

    err = os.WriteFile(lat_path, []byte("{ \"latest\": \"1\" }"), 0644)
    if err != nil {
        t.Fatalf("failed to update the ..latest file; %v", err)
    }

    t.Run("reset latest", func(t *testing.T) {
        changed, err := ignoreNonLatest(registry, project, asset)
        if err != nil {
            t.Fatalf("unexpected errors when ignoring non-latest version; %v", err)
        }
        if !changed {
            t.Fatal("expected some kind of change when ignoring non-latest version")
        }

        if _, err := os.Stat(filepath.Join(asset_dir, "1", ".SewerRatignore")); !errors.Is(err, os.ErrNotExist) {
            t.Fatalf("expected no .SewerRatignore file for version 1")
        }
        if _, err := os.Stat(filepath.Join(asset_dir, "2", ".SewerRatignore")); err != nil {
            t.Fatalf("expected a .SewerRatignore file for version 2")
        }
        if _, err := os.Stat(filepath.Join(asset_dir, "3", ".SewerRatignore")); err != nil {
            t.Fatalf("expected a .SewerRatignore file for version 3")
        }
    })

    err = os.Remove(lat_path)
    if err != nil {
        t.Fatalf("failed to remove the ..latest file; %v", err)
    }

    t.Run("no latest", func(t *testing.T) {
        changed, err := ignoreNonLatest(registry, project, asset)
        if err != nil {
            t.Fatalf("unexpected errors when ignoring non-latest version; %v", err)
        }
        if !changed {
            t.Fatal("expected some kind of change when ignoring non-latest version")
        }

        if _, err := os.Stat(filepath.Join(asset_dir, "1", ".SewerRatignore")); err != nil {
            t.Fatalf("expected a .SewerRatignore file for version 1")
        }
        if _, err := os.Stat(filepath.Join(asset_dir, "2", ".SewerRatignore")); err != nil {
            t.Fatalf("expected a .SewerRatignore file for version 2")
        }
        if _, err := os.Stat(filepath.Join(asset_dir, "3", ".SewerRatignore")); err != nil {
            t.Fatalf("expected a .SewerRatignore file for version 3")
        }
    })
}
