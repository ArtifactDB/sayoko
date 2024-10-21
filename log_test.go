package main

import (
    "testing"
    "os"
    "time"
    "path/filepath"
)

func TestReadLog(t *testing.T) {
    logdir, err := os.MkdirTemp("", "")
    if err != nil {
        t.Fatalf("failed to create log directory; %v", err)
    }

    log_path := filepath.Join(logdir, time.Now().Format(time.RFC3339) + "_111111")
    err = os.WriteFile(log_path, []byte("{ \"action\": \"add-version\", \"project\": \"foo\", \"asset\": \"bar\" }"), 0644)
    if err != nil {
        t.Fatalf("failed to create a new log file; %v", err)
    }

    loginfo, err := readLog(log_path)
    if err != nil {
        t.Fatalf("failed to read the log file; %v", err)
    }
    if loginfo.Action != "add-version" || loginfo.Project != "foo" || loginfo.Asset != "bar" {
        t.Fatal("unexpected contents of the log file")
    }
}

func TestCheckLogs(t *testing.T) {
    registry, err := os.MkdirTemp("", "")
    if err != nil {
        t.Fatalf("failed to create registry; %v", err)
    }

    // Mocking up logs.
    logdir := filepath.Join(registry, "..logs")
    err = os.Mkdir(logdir, 0755)
    if err != nil {
        t.Fatalf("failed to create a new log directory; %v", err)
    }

    log_path := filepath.Join(logdir, "2022-02-22T02:22:22Z_111111")
    err = os.WriteFile(log_path, []byte("{ \"action\": \"add-version\", \"project\": \"foo\", \"asset\": \"bar\", \"version\": \"1\" }"), 0644)
    if err != nil {
        t.Fatalf("failed to create a new log file; %v", err)
    }

    log_path = filepath.Join(logdir, "2023-03-23T02:22:22Z_111111")
    err = os.WriteFile(log_path, []byte("{ \"action\": \"delete-version\", \"project\": \"foo\", \"asset\": \"whee\", \"version\": \"1\" }"), 0644)
    if err != nil {
        t.Fatalf("failed to create a new log file; %v", err)
    }

    log_path = filepath.Join(logdir, "2024-04-24T02:22:22Z_111111")
    err = os.WriteFile(log_path, []byte("{ \"action\": \"delete-asset\", \"project\": \"shibuya\", \"asset\": \"kanon\" }"), 0644)
    if err != nil {
        t.Fatalf("failed to create a new log file; %v", err)
    }

    // Creating project directories.
    err = os.Mkdir(filepath.Join(registry, "foo"), 0755)
    if err != nil {
        t.Fatalf("failed to create a 'foo' project; %v", err)
    }

    err = os.Mkdir(filepath.Join(registry, "shibuya"), 0755)
    if err != nil {
        t.Fatalf("failed to create a 'shibuya' project; %v", err)
    }

    url := getSewerRatUrl()
    last_scan, err := time.Parse(time.RFC3339, "2021-01-21T02:22:22Z")
    if err != nil {
        t.Fatalf("failed to parse time; %v", err)
    }

    t.Run("no registered", func(t *testing.T) {
        to_reignore := map[string]bool{}
        to_reregister := map[string]bool{}

        output, err := checkLogs(registry, url, to_reignore, to_reregister, last_scan)
        if err != nil {
            t.Fatalf("failed to check logs; %v", err)
        }
        if output.Year() != 2024 {
            t.Fatal("expected the latest timestamp to be reported")
        }

        if len(to_reignore) > 0 || len(to_reregister) > 0 {
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
        to_reregister := map[string]bool{}

        output, err := checkLogs(registry, url, to_reignore, to_reregister, last_scan)
        if err != nil {
            t.Fatalf("failed to check logs; %v", err)
        }
        if output.Year() != 2024 {
            t.Fatal("expected the latest timestamp to be reported")
        }

        if _, ok := to_reignore["foo/bar"]; !ok {
            t.Fatal("expected 'foo/bar' to be reignored")
        }
        if _, ok := to_reignore["foo/whee"]; !ok {
            t.Fatal("expected 'foo/bar' to be reignored")
        }
        if _, ok := to_reregister["shibuya"]; !ok {
            t.Fatal("expected 'shibuya' to be reregistered")
        }

        to_reignore = map[string]bool{}
        to_reregister = map[string]bool{}
        output2, err := checkLogs(registry, url, to_reignore, to_reregister, output)
        if err != nil {
            t.Fatalf("failed to check logs; %v", err)
        }
        if output != output2 {
            t.Fatal("expected the latest timestamp to be unchanged")
        }
        if len(to_reignore) > 0 || len(to_reregister) > 0 {
            t.Fatal("expected no action for stale logs")
        }
    })

    // Trying a more recent time.
    t.Run("more recent time stemp", func(t *testing.T) {
        to_reignore := map[string]bool{}
        to_reregister := map[string]bool{}
        last_scan := time.Now()

        output, err := checkLogs(registry, url, to_reignore, to_reregister, last_scan)
        if err != nil {
            t.Fatalf("failed to check logs; %v", err)
        }
        if output != last_scan {
            t.Fatal("expected the latest timestamp to be unchanged")
        }

        if len(to_reignore) > 0 || len(to_reregister) > 0 {
            t.Fatal("expected no action for old logs")
        }
    })
}
