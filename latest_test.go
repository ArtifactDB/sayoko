package main

import (
    "os"
    "path/filepath"
    "testing"
    "strings"
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
    if latest.Version != "" {
        t.Fatalf("..latest file should be absent; %v", err)
    }

    err = os.WriteFile(lat_path, []byte("{ \"version\": \"foobar\" }"), 0644)
    if err != nil {
        t.Fatalf("failed to write to the ..latest file; %v", err)
    }

    latest, err = readLatestFile(lat_path)
    if err != nil {
        t.Fatalf("failed to read the ..latest file; %v", err)
    }
    if latest.Version != "foobar" {
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
    err = os.WriteFile(lat_path, []byte("{ \"version\": \"3\" }"), 0644)
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

    url := getSewerRatUrl()

    // Simple initial run.
    {
        err := ignoreNonLatest(url, asset_dir, false)
        if err != nil {
            t.Fatal(err)
        }

        found, err := listRegisteredSubdirectories(url, asset_dir)
        if err != nil {
            t.Fatal(err)
        }

        if len(found) != 1 || found[0] != "3" {
            t.Errorf("expected only version '3' to be registered; %v", found)
        }
    }

    // Reset the latest.
    {
        err := os.WriteFile(lat_path, []byte("{ \"version\": \"1\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to update the ..latest file; %v", err)
        }

        err = ignoreNonLatest(url, asset_dir, false)
        if err != nil {
            t.Fatal(err)
        }

        found, err := listRegisteredSubdirectories(url, asset_dir)
        if err != nil {
            t.Fatal(err)
        }

        if len(found) != 1 || found[0] != "1" {
            t.Errorf("expected only version '1' to be registered; %v", found)
        }
    }

    // No latest file.
    {
        err = os.Remove(lat_path)
        if err != nil {
            t.Fatalf("failed to remove the ..latest file; %v", err)
        }

        err := ignoreNonLatest(url, asset_dir, false)
        if err != nil {
            t.Fatal(err)
        }

        found, err := listRegisteredSubdirectories(url, asset_dir)
        if err != nil {
            t.Fatal(err)
        }

        if len(found) != 0 {
            t.Errorf("expected no version to be registered; %v", found)
        }
    }

    // Forcibly reregstering.
    {
        err := os.WriteFile(lat_path, []byte("{ \"version\": \"3\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to write to the ..latest file; %v", err)
        }

        err = ignoreNonLatest(url, asset_dir, false)
        if err != nil {
            t.Fatal(err)
        }
        found, err := listRegisteredSubdirectories(url, asset_dir)
        if err != nil {
            t.Fatal(err)
        }
        if len(found) != 1 || found[0] != "3" {
            t.Errorf("expected a single version to be registered; %v", found)
        }

        // We shouldn't get a registration error if we don't force the issue.
        version_dir := filepath.Join(asset_dir, "3")
        err = os.RemoveAll(version_dir)
        if err != nil {
            t.Fatal(err)
        }

        err = ignoreNonLatest(url, asset_dir, false)
        if err != nil {
            t.Fatal(err)
        }
        found, err = listRegisteredSubdirectories(url, asset_dir)
        if err != nil {
            t.Fatal(err)
        }
        if len(found) != 1 || found[0] != "3" {
            t.Errorf("expected a single version to be registered; %v", found)
        }

        // But if we do force it, we should see an error because the directory doesn't exist.
        err = ignoreNonLatest(url, asset_dir, true)
        if err == nil || !strings.Contains(err.Error(), "does not exist") {
            t.Error("expected an error from forced reregistration")
        }
    }
}
