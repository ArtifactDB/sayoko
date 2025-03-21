package main

import (
    "testing"
    "os"
    "sort"
    "path/filepath"
    "encoding/json"
    "fmt"
    "bytes"
    "net/http"
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
    names := []string{ "metadata.json" }
    err = registerDirectory(url, dir1, names)
    if err != nil {
        return nil, err
    }
    err = registerDirectory(url, dir2, names)
    if err != nil {
        return nil, err
    }
    err = registerDirectory(url, dir3, names)
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

func TestRegisterDirectoryNames(t *testing.T) {
    dir, err := os.MkdirTemp("", "")
    if err != nil {
        t.Fatal(err)
    }

    err = os.WriteFile(filepath.Join(dir, "foo.json"), []byte("{ \"first\": \"athena\", \"last\": \"glory\", \"show\": \"aria the animation\" }"), 0644)
    if err != nil {
        t.Fatal(err)
    }
    err = os.WriteFile(filepath.Join(dir, "metadata.json"), []byte("{ \"series\": \"aria\", \"season\": 3 }"), 0644)
    if err != nil {
        t.Fatal(err)
    }

    url := getSewerRatUrl()
    defer deregisterDirectory(url, dir) // to avoid affecting other tests.

    querySewerRat := func(query string) ([]string, error) {
        b, err := json.Marshal(map[string]interface{}{ "type": "text", "text": query })
        if err != nil {
            return nil, fmt.Errorf("failed to create query request body; %w", err)
        }

        r := bytes.NewReader(b)
        resp, err := http.Post(url + "/query", "application/json", r)
        if err != nil {
            return nil, fmt.Errorf("failed to request query; %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 300 {
            err := parseFailure(resp)
            return nil, fmt.Errorf("failed query; %w", err)
        }

        type queryHit struct {
            Path string `json:"path"`
        }
        type allQueryHits struct {
            Results []queryHit `json:"results"`
        }
        dec := json.NewDecoder(resp.Body)
        decoded := allQueryHits{}
        err = dec.Decode(&decoded)
        if err != nil {
            return nil, fmt.Errorf("failed to parse query response; %w", err) 
        }

        output := []string{}
        for _, q := range decoded.Results {
            output = append(output, q.Path)
        }
        return output, nil
    }

    err = registerDirectory(url, dir, []string{ "foo.json" })
    if err != nil {
        t.Fatal(err)
    }
    output, err := querySewerRat("athena")
    if err != nil {
        t.Fatal(err)
    }
    if len(output) != 1 || filepath.Base(output[0]) != "foo.json" {
        t.Errorf("unexpected query result when registering foo.json; %v", output)
    }

    err = registerDirectory(url, dir, []string{ "foo.json", "metadata.json" })
    if err != nil {
        t.Fatal(err)
    }
    output, err = querySewerRat("aria")
    if err != nil {
        t.Fatal(err)
    }
    sort.Strings(output)
    if len(output) != 2 || filepath.Base(output[0]) != "foo.json" || filepath.Base(output[1]) != "metadata.json" {
        t.Errorf("unexpected query result when registering both foo.json and metadata.json; %v", output)
    }
}
