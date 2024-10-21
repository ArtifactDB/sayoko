package main

import (
    "testing"
    "os"
    "path/filepath"
)

func TestFullScan(t *testing.T) {
    registry, err := os.MkdirTemp("", "")
    if err != nil {
        t.Fatalf("failed to create registry; %v", err)
    }

    // Creating project and asset directories.
    err = os.Mkdir(filepath.Join(registry, "foo"), 0755)
    if err != nil {
        t.Fatalf("failed to create a 'foo' project; %v", err)
    }

    err = os.Mkdir(filepath.Join(registry, "foo", "bar"), 0755)
    if err != nil {
        t.Fatalf("failed to create a 'foo/bar' project; %v", err)
    }

    err = os.Mkdir(filepath.Join(registry, "shibuya"), 0755)
    if err != nil {
        t.Fatalf("failed to create a 'shibuya' project; %v", err)
    }

    err = os.Mkdir(filepath.Join(registry, "shibuya", "kanon"), 0755)
    if err != nil {
        t.Fatalf("failed to create a 'shibuya/kanon' project; %v", err)
    }

    url := getSewerRatUrl()

    t.Run("no registered", func(t *testing.T) {
        to_reignore := map[string]bool{}

        err := fullScan(registry, url, to_reignore)
        if err != nil {
            t.Fatalf("failed to check logs; %v", err)
        }

        if len(to_reignore) > 0 {
            t.Fatal("expected no action for unregistered projects")
        }
    })

    // Now registering everything.
    err = reregisterProject(url, filepath.Join(registry, "foo"))
    if err != nil {
        t.Fatalf("failed to register the 'foo' project; %v", err)
    }

    err = reregisterProject(url, filepath.Join(registry, "shibuya"))
    if err != nil {
        t.Fatalf("failed to register the 'shibuya' project; %v", err)
    }

    t.Run("all registered", func(t *testing.T) {
        to_reignore := map[string]bool{}

        err := fullScan(registry, url, to_reignore)
        if err != nil {
            t.Fatalf("failed to check logs; %v", err)
        }

        if _, ok := to_reignore["foo/bar"]; !ok {
            t.Fatal("expected 'foo/bar' to be reignored")
        }
        if _, ok := to_reignore["shibuya/kanon"]; !ok {
            t.Fatal("expected 'shibuya/kanon' to be reignored")
        }
    })
}
