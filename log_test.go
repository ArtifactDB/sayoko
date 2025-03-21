package main

import (
    "testing"
    "os"
    "time"
    "path/filepath"
    "sort"
    "strings"
)

func TestReadLog(t *testing.T) {
    logdir, err := os.MkdirTemp("", "")
    if err != nil {
        t.Fatalf("failed to create log directory; %v", err)
    }

    log_path := filepath.Join(logdir, time.Now().Format(time.RFC3339) + "_111111")
    err = os.WriteFile(log_path, []byte("{ \"type\": \"add-version\", \"project\": \"foo\", \"asset\": \"bar\" }"), 0644)
    if err != nil {
        t.Fatalf("failed to create a new log file; %v", err)
    }

    loginfo, err := readLog(log_path)
    if err != nil {
        t.Fatalf("failed to read the log file; %v", err)
    }
    if loginfo.Type != "add-version" || loginfo.Project != "foo" || loginfo.Asset != "bar" {
        t.Fatal("unexpected contents of the log file")
    }
}

func TestProcessLogs(t *testing.T) {
    registry, err := os.MkdirTemp("", "")
    if err != nil {
        t.Fatalf("failed to create registry; %v", err)
    }

    // Creating project directories.
    err = os.MkdirAll(filepath.Join(registry, "foo", "bar", "1"), 0755)
    if err != nil {
        t.Fatalf("failed to create a 'foo/bar/1' project; %v", err)
    }
    err = os.WriteFile(filepath.Join(registry, "foo", "bar", "..latest"), []byte("{ \"version\": \"1\" }"), 0644)
    if err != nil {
        t.Fatal(err)
    }

    err = os.MkdirAll(filepath.Join(registry, "shibuya", "kanon", "1"), 0755)
    if err != nil {
        t.Fatalf("failed to create a 'shibuya/kanon/2' project; %v", err)
    }
    err = os.MkdirAll(filepath.Join(registry, "shibuya", "kanon", "2"), 0755)
    if err != nil {
        t.Fatalf("failed to create a 'shibuya/kanon/2' project; %v", err)
    }
    err = os.WriteFile(filepath.Join(registry, "shibuya", "kanon", "..latest"), []byte("{ \"version\": \"2\" }"), 0644)
    if err != nil {
        t.Fatal(err)
    }

    // Creating the log directory.
    logdir := filepath.Join(registry, "..logs")
    err = os.Mkdir(logdir, 0755)
    flushLogs := func() {
        listing, err := os.ReadDir(logdir)
        if err != nil {
            t.Fatal(err)
        }
        for _, entry := range listing {
            err := os.Remove(filepath.Join(logdir, entry.Name()))
            if err != nil {
                t.Fatal(err)
            }
        }
    }

    last_scan, err := time.Parse(time.RFC3339, "2021-01-21T02:22:22Z")
    if err != nil {
        t.Fatalf("failed to parse time; %v", err)
    }

    url := getSewerRatUrl()
    names := []string{ "metadata.json" }

    // Adding a new version.
    {
        flushLogs()

        log_path := filepath.Join(logdir, "2022-02-22T02:22:22Z_111111")
        err := os.WriteFile(log_path, []byte("{ \"type\": \"add-version\", \"project\": \"foo\", \"asset\": \"bar\", \"version\": \"1\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to create a new log file; %v", err)
        }

        log_path = filepath.Join(logdir, "2023-03-23T03:33:33Z_111111")
        err = os.WriteFile(log_path, []byte("{ \"type\": \"add-version\", \"project\": \"shibuya\", \"asset\": \"kanon\", \"version\": \"2\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to create a new log file; %v", err)
        }

        _, err = processLogs(url, registry, names, last_scan)
        if err != nil {
            t.Fatal(err)
        }

        found, err := listRegisteredSubdirectories(url, registry)
        if err != nil {
            t.Fatal(err)
        }
        sort.Strings(found)
        if len(found) != 2 || found[0] != "foo/bar/1" || found[1] != "shibuya/kanon/2" {
            t.Errorf("unexpected result from registering from the logs; %v", found)
        }
    }

    // Deleting a version.
    {
        flushLogs()

        whee_path := filepath.Join(registry, "foo", "whee", "4")
        err := os.MkdirAll(whee_path, 0755)
        if err != nil {
            t.Fatal(err)
        }
        err = registerDirectory(url, whee_path, names)
        if err != nil {
            t.Fatal(err)
        }
        err = os.RemoveAll(whee_path)
        if err != nil {
            t.Fatal(err)
        }

        // Confirming that we registered it successfully.
        found, err := listRegisteredSubdirectories(url, filepath.Join(registry, "foo", "whee"))
        if err != nil {
            t.Fatal(err)
        }
        if len(found) != 1 || found[0] != "4" {
            t.Fatal("expected 'foo/whee/4' to be registered")
        }

        log_path := filepath.Join(logdir, "2022-02-22T02:22:22Z_111111")
        err = os.WriteFile(log_path, []byte("{ \"type\": \"delete-version\", \"project\": \"foo\", \"asset\": \"whee\", \"version\": \"4\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to create a new log file; %v", err)
        }

        _, err = processLogs(url, registry, names, last_scan)
        if err != nil {
            t.Fatal(err)
        }

        found, err = listRegisteredSubdirectories(url, filepath.Join(registry, "foo"))
        if err != nil {
            t.Fatal(err)
        }
        if len(found) != 1 || found[0] != "bar/1" {
            t.Errorf("unexpected result from registering from the logs; %v", found)
        }
    }

    // Deleting an asset. 
    {
        flushLogs()

        for _, ver := range []string{ "1", "2", "3" } {
            whee_path := filepath.Join(registry, "shibuya", "aria", ver)
            err := os.MkdirAll(whee_path, 0755)
            if err != nil {
                t.Fatal(err)
            }
            err = registerDirectory(url, whee_path, names)
            if err != nil {
                t.Fatal(err)
            }
        }
        err = os.RemoveAll(filepath.Join(registry, "shibuya", "aria"))
        if err != nil {
            t.Fatal(err)
        }

        // Confirming that we registered it successfully.
        found, err := listRegisteredSubdirectories(url, filepath.Join(registry, "shibuya", "aria"))
        if err != nil {
            t.Fatal(err)
        }
        if len(found) != 3 {
            t.Fatal("expected 'shibuya/aria' to have multiple registered versions")
        }

        log_path := filepath.Join(logdir, "2022-02-22T02:22:22Z_111111")
        err = os.WriteFile(log_path, []byte("{ \"type\": \"delete-asset\", \"project\": \"shibuya\", \"asset\": \"aria\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to create a new log file; %v", err)
        }

        _, err = processLogs(url, registry, names, last_scan)
        if err != nil {
            t.Fatal(err)
        }

        found, err = listRegisteredSubdirectories(url, filepath.Join(registry, "shibuya"))
        if err != nil {
            t.Fatal(err)
        }
        if len(found) != 1 || found[0] != "kanon/2" {
            t.Errorf("unexpected result from registering from the logs; %v", found)
        }
    }

    // Deleting a project.
    {
        flushLogs()

        whee_path := filepath.Join(registry, "heanna", "sumire", "4")
        err := os.MkdirAll(whee_path, 0755)
        if err != nil {
            t.Fatal(err)
        }
        err = registerDirectory(url, whee_path, names)
        if err != nil {
            t.Fatal(err)
        }
        err = os.RemoveAll(whee_path)
        if err != nil {
            t.Fatal(err)
        }

        // Confirming that we registered it successfully.
        found, err := listRegisteredSubdirectories(url, filepath.Join(registry, "heanna"))
        if err != nil {
            t.Fatal(err)
        }
        if len(found) != 1 || found[0] != "sumire/4" {
            t.Fatalf("expected 'shibuya/aria' to have multiple registered versions; %v", found)
        }

        log_path := filepath.Join(logdir, "2022-02-22T02:22:22Z_111111")
        err = os.WriteFile(log_path, []byte("{ \"type\": \"delete-project\", \"project\": \"heanna\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to create a new log file; %v", err)
        }

        _, err = processLogs(url, registry, names, last_scan)
        if err != nil {
            t.Fatal(err)
        }

        found, err = listRegisteredSubdirectories(url, registry)
        if err != nil {
            t.Fatal(err)
        }
        sort.Strings(found)
        if len(found) != 2 || found[0] != "foo/bar/1" || found[1] != "shibuya/kanon/2" {
            t.Errorf("unexpected result from registering from the logs; %v", found)
        }
    }

    // Assorted failures.
    {
        flushLogs()
        log_path := filepath.Join(logdir, "2022-02-22T02:22:22Z_111111")

        err := os.WriteFile(log_path, []byte("{ \"type\": \"delete-version\", \"asset\": \"whee\", \"version\": \"4\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to create a new log file; %v", err)
        }
        _, err = processLogs(url, registry, names, last_scan)
        if err == nil || !strings.Contains(err.Error(), "empty") {
            t.Error("lack of error when project field is empty")
        }

        err = os.WriteFile(log_path, []byte("{ \"type\": \"delete-version\", \"project\": \"foo\", \"version\": \"4\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to create a new log file; %v", err)
        }
        _, err = processLogs(url, registry, names, last_scan)
        if err == nil || !strings.Contains(err.Error(), "empty") {
            t.Error("lack of error when asset field is empty")
        }

        err = os.WriteFile(log_path, []byte("{ \"type\": \"delete-asset\", \"project\": \"whee\", \"version\": \"4\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to create a new log file; %v", err)
        }
        _, err = processLogs(url, registry, names, last_scan)
        if err == nil || !strings.Contains(err.Error(), "empty") {
            t.Error("lack of error when asset field is empty")
        }

        err = os.WriteFile(log_path, []byte("{ \"type\": \"delete-project\", \"version\": \"4\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to create a new log file; %v", err)
        }
        _, err = processLogs(url, registry, names, last_scan)
        if err == nil || !strings.Contains(err.Error(), "empty") {
            t.Error("lack of error when project field is empty")
        }
    }

    // Respects timestamps.
    {
        flushLogs()
        err = deregisterAllSubdirectories(url, registry)
        if err != nil {
            t.Fatal(err)
        }

        whee_path := filepath.Join(registry, "heanna", "sumire", "4")
        err := os.MkdirAll(whee_path, 0755)
        if err != nil {
            t.Fatal(err)
        }
        err = registerDirectory(url, whee_path, names)
        if err != nil {
            t.Fatal(err)
        }
        err = os.RemoveAll(whee_path)
        if err != nil {
            t.Fatal(err)
        }

        // Confirming that we registered it successfully.
        found, err := listRegisteredSubdirectories(url, registry)
        if err != nil {
            t.Fatal(err)
        }
        if len(found) != 1 || found[0] != "heanna/sumire/4" {
            t.Fatalf("expected only 'heanna/sumire/4' to be registered; %v", found)
        }

        // Processing multiple logs.
        log_path := filepath.Join(logdir, "2022-02-22T02:22:22Z_111111")
        err = os.WriteFile(log_path, []byte("{ \"type\": \"add-version\", \"project\": \"foo\", \"asset\": \"bar\", \"version\": \"1\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to create a new log file; %v", err)
        }

        log_path = filepath.Join(logdir, "2024-04-24T04:44:44Z_111111")
        err = os.WriteFile(log_path, []byte("{ \"type\": \"delete-version\", \"project\": \"heanna\", \"asset\": \"sumire\", \"version\": \"4\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to create a new log file; %v", err)
        }

        log_path = filepath.Join(logdir, "2020-02-22T02:22:22Z_111111")
        err = os.WriteFile(log_path, []byte("{ \"type\": \"add-version\", \"project\": \"shibuya\", \"asset\": \"kanon\", \"version\": \"2\" }"), 0644)
        if err != nil {
            t.Fatalf("failed to create a new log file; %v", err)
        }

        new_time, err := processLogs(url, registry, names, last_scan)
        if err != nil {
            t.Fatal(err)
        }
        if new_time.Year() != 2024 {
            t.Error("unexpected year from the latest timestamp")
        }

        // Confirming that we only updated the things past the last_scan timestamp.
        found, err = listRegisteredSubdirectories(url, registry)
        if err != nil {
            t.Fatal(err)
        }
        if len(found) != 1 || found[0] != "foo/bar/1" {
            t.Fatalf("expected only 'foo/bar/1' to be registered; %v", found)
        }
    }
}
