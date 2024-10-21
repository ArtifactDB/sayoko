package main

import (
    "testing"
    "os"
    "path/filepath"
)

func getSewerRatUrl() string {
    url := os.Getenv("SEWERRAT_URL")
    if url == "" {
        url = "http://0.0.0.0:8080"
    }
    return url
}

func TestIsProjectRegistered(t *testing.T) {
    registry, err := os.MkdirTemp("", "")
    if err != nil {
        t.Fatalf("failed to create temporary directory; %v", err)
    }

    registry, err = filepath.Abs(registry)
    if err != nil {
        t.Fatalf("failed to convert registry into an absolute path; %v", err)
    }

    project := "liella"
    project_dir := filepath.Join(registry, project)
    err = os.Mkdir(project_dir, 0755)
    if err != nil {
        t.Fatalf("failed to create the project directory; %v", err)
    }

    url := getSewerRatUrl()
    err = reregisterProject(url, project_dir)
    if err != nil {
        t.Fatalf("failed to register a project directory; %v", err)
    }

    t.Run("basic", func(t *testing.T) {
        is_reg, err := isProjectRegistered(registry, url, project)
        if err != nil {
            t.Fatalf("failed to check if the project is registered; %v", err)
        }
        if !is_reg {
            t.Fatal("expected the project to be registered")
        }

        project2 := "aquors"
        project2_dir := filepath.Join(registry, project2)
        err = os.Mkdir(project2_dir, 0755)
        if err != nil {
            t.Fatalf("failed to create another project directory; %v", err)
        }

        is_reg, err = isProjectRegistered(registry, url, project2)
        if err != nil {
            t.Fatalf("failed to check if the project is registered; %v", err)
        }
        if is_reg {
            t.Fatal("expected the project to not be registered")
        }
    })

    t.Run("cached", func(t *testing.T) {
        cache := map[string]bool{}

        is_reg, err := isProjectRegisteredWithCache(registry, url, project, cache)
        if err != nil {
            t.Fatalf("failed to check if the project is registered; %v", err)
        }
        if !is_reg {
            t.Fatal("expected the project to be registered")
        }

        cache[project] = false
        is_reg, err = isProjectRegisteredWithCache(registry, url, project, cache)
        if err != nil {
            t.Fatalf("failed to check if the project is registered; %v", err)
        }
        if is_reg {
            t.Fatal("expected the project to be non-registered")
        }
    })

}
