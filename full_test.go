package main

import (
    "testing"
    "os"
    "path/filepath"
    "sort"
)

func TestFullScan(t *testing.T) {
    registry, err := os.MkdirTemp("", "")
    if err != nil {
        t.Fatalf("failed to create registry; %v", err)
    }

    // Creating project and asset directories.
    err = os.MkdirAll(filepath.Join(registry, "foo", "bar", "1"), 0755)
    if err != nil {
        t.Fatalf("failed to create a 'foo/bar' project; %v", err)
    }
    err = os.WriteFile(filepath.Join(registry, "foo", "bar", "..latest"), []byte("{ \"version\": \"1\" }"), 0644)
    if err != nil {
        t.Fatal(err)
    }

    err = os.MkdirAll(filepath.Join(registry, "shibuya", "kanon", "2"), 0755)
    if err != nil {
        t.Fatalf("failed to create a 'shibuya/kanon' project; %v", err)
    }
    err = os.WriteFile(filepath.Join(registry, "shibuya", "kanon", "..latest"), []byte("{ \"version\": \"2\" }"), 0644)
    if err != nil {
        t.Fatal(err)
    }

    err = os.MkdirAll(filepath.Join(registry, "shibuya", "aria", "1"), 0755)
    if err != nil {
        t.Fatalf("failed to create a 'shibuya/kanon' project; %v", err)
    }
    err = os.WriteFile(filepath.Join(registry, "shibuya", "aria", "..latest"), []byte("{ \"version\": \"1\" }"), 0644)
    if err != nil {
        t.Fatal(err)
    }

    url := getSewerRatUrl()

    // Initial run registers everything.
    {
        err := fullScan(url, registry)
        if err != nil {
            t.Fatal(err)
        }

        found, err := listRegisteredSubdirectories(url, registry)
        if err != nil {
            t.Fatal(err)
        }

        sort.Strings(found)
        if len(found) != 3 || found[0] != "foo/bar/1" || found[1] != "shibuya/aria/1" || found[2] != "shibuya/kanon/2" {
            t.Errorf("unexpected results after a full scan; %v", found)
        }
    }

    // Next run deregisters all the shibuya entries.
    {
        err := os.RemoveAll(filepath.Join(registry, "shibuya"))
        if err != nil {
            t.Fatal(err)
        }

        err = fullScan(url, registry)
        if err != nil {
            t.Fatal(err)
        }

        found, err := listRegisteredSubdirectories(url, registry)
        if err != nil {
            t.Fatal(err)
        }

        if len(found) != 1 || found[0] != "foo/bar/1" {
            t.Errorf("unexpected results after a full scan; %v", found)
        }
    }
}
