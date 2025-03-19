package main

import (
    "testing"
    "os"
    "sort"
    "path/filepath"
)

func getSewerRatUrl() string {
    url := os.Getenv("SEWERRAT_URL")
    if url == "" {
        url = "http://0.0.0.0:8080"
    }
    return url
}

func setupDirectories() ([]string, error) {
    dir, err := os.MkdirTemp("", "")
    if err != nil {
        return nil, err
    }

    dir1 := filepath.Join(dir, "FOO")
    err = os.Mkdir(dir1, 0755)
    if err != nil {
        return nil, err
    }

    dir2 := filepath.Join(dir, "BAR")
    err = os.Mkdir(dir2, 0755)
    if err != nil {
        return nil, err
    }

    altdir, err := os.MkdirTemp("", "")
    if err != nil {
        return nil, err
    }

    dir3 := filepath.Join(altdir, "WHEE")
    err = os.Mkdir(dir3, 0755)
    if err != nil {
        return nil, err
    }

    url := getSewerRatUrl()
    err = registerDirectory(url, dir1, true)
    if err != nil {
        return nil, err
    }
    err = registerDirectory(url, dir2, true)
    if err != nil {
        return nil, err
    }
    err = registerDirectory(url, dir3, true)
    if err != nil {
        return nil, err
    }

    return []string{ dir1, dir2, dir3 }, nil
}

func TestListRegisteredSubdirectories(t *testing.T) {
    dirs, err := setupDirectories()
    if err != nil {
        t.Fatal(err)
    }
    dir1 := dirs[0]
    dir3 := dirs[2]

    url := getSewerRatUrl()
    found, err := listRegisteredSubdirectories(url, filepath.Dir(dir1))
    if err != nil {
        t.Fatal(err)
    }
    sort.Strings(found)
    if len(found) != 2 || found[0] != "BAR" || found[1] != "FOO" {
        t.Errorf("unexpected results from listing; %v", found)
    }

    found, err = listRegisteredSubdirectories(url, filepath.Dir(dir3))
    if err != nil {
        t.Fatal(err)
    }
    if len(found) != 1 || found[0] != "WHEE" { // dir1 is its own registered entity.
        t.Errorf("unexpected results from listing; %v", found)
    }
}

func TestDeregisterAllSubdirectories(t *testing.T) {
    dirs, err := setupDirectories()
    if err != nil {
        t.Fatal(err)
    }
    dir1 := dirs[0]
    dir3 := dirs[2]

    url := getSewerRatUrl()
    err = deregisterAllSubdirectories(url, filepath.Dir(dir1))
    if err != nil {
        t.Fatal(err)
    }

    found, err := listRegisteredSubdirectories(url, filepath.Dir(dir1))
    if err != nil {
        t.Fatal(err)
    }
    if len(found) > 0 {
        t.Errorf("all directories should have been deregistered; %v", found)
    }

    // Other subdirectories are still okay.
    found, err = listRegisteredSubdirectories(url, filepath.Dir(dir3))
    if err != nil {
        t.Fatal(err)
    }
    if len(found) != 1 || found[0] != "WHEE" { // dir1 is its own registered entity.
        t.Errorf("unexpected results from listing; %v", found)
    }
}

func TestDeregisterMissingSubdirectories(t *testing.T) {
    dirs, err := setupDirectories()
    if err != nil {
        t.Fatal(err)
    }
    dir1 := dirs[0]
    dir2 := dirs[1]

    err = os.RemoveAll(dir2)
    if err != nil {
        t.Fatal(err)
    }

    url := getSewerRatUrl()
    err = deregisterMissingSubdirectories(url, filepath.Dir(dir1))
    if err != nil {
        t.Fatal(err)
    }

    found, err := listRegisteredSubdirectories(url, filepath.Dir(dir1))
    if err != nil {
        t.Fatal(err)
    }
    if len(found) != 1 || found[0] != "FOO" {
        t.Errorf("only missing directories should have been deregistered; %v", found)
    }
}
