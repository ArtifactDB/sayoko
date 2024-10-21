package main

import (
    "os"
    "flag"
    "log"
    "fmt"
    "time"
    "sync"
    "strings"
    "path/filepath"
    "errors"
)

const last_scan_path = ".sayoko_last_scan"

func retrieveLastScanTime() time.Time {
    last_scan_raw, err := os.ReadFile(last_scan_path)
    if err == nil {
        candidate, err := time.Parse(time.RFC3339, string(last_scan_raw))
        if err == nil {
            return candidate
        } else {
            log.Printf("failed to parse the last scan time; %v", err)
        }
    } else if !errors.Is(err, os.ErrNotExist) {
        log.Printf("failed to read the last scan time; %v", err)
    }
    return time.Now()
}

func depositLastScanTime(last_scan time.Time) {
    err := os.WriteFile(last_scan_path, []byte(last_scan.Format(time.RFC3339)), 0644)
    if err != nil {
        log.Printf("failed to write the last scan time; %v", err)
    }
}

func unpackKey(key string) (string, string) {
    pos := strings.IndexByte(key, '/')
    return key[:pos], key[(pos + 1):]
}

func main() {
    gpath := flag.String("registry", "", "Path to the gobbler registry")
    surl := flag.String("url", "", "URL of the SewerRat instance")
    log_time := flag.Int("log", 10, "Interval in which to check for new logs, in minutes")
    full_time := flag.Int("full", 24, "Interval in which to do a full check, in hours")
    flag.Parse()

    registry := *gpath
    rest_url := *surl
    if registry == "" || rest_url == "" {
        flag.Usage()
        os.Exit(1)
    }

    if !filepath.IsAbs(registry) {
        fmt.Println("expected an absolute file path for the registry")
        os.Exit(1)
    }

    to_reignore := map[string]bool{}
    to_reregister := map[string]bool{}
    var lock sync.Mutex
    ch_reignore := make(chan bool)
    ch_reregister := make(chan bool)

    // Timer to inspect logs.
    go func() {
        last_scan := retrieveLastScanTime()
        timer := time.NewTicker(time.Minute * time.Duration(*log_time))
        for {
            <-timer.C
            lock.Lock()
            new_last_scan, err := checkLogs(registry, rest_url, to_reignore, to_reregister, last_scan)
            lock.Unlock()
            if err != nil {
                log.Printf("detected failures for log check; %v", err)
            }
            last_scan = new_last_scan // this can be set regardless of 'err'.
            depositLastScanTime(last_scan)
            ch_reignore <- true
        }
    }()

    // Timer to scan the entire registry.
    go func() {
        timer := time.NewTicker(time.Hour * time.Duration(*full_time))
        for {
            <-timer.C
            lock.Lock()
            err := fullScan(registry, rest_url, to_reignore)
            lock.Unlock()
            if err != nil {
                log.Printf("detected failures for log check; %v", err)
            }
            ch_reignore <- true
        }
    }()

    // Listener to update the ignore files.
    go func() {
        for {
            <- ch_reignore
            any_changed := false
            lock.Lock()
            for k, _ := range to_reignore {
                project, asset := unpackKey(k)
                changed, err := ignoreNonLatest(registry, project, asset)
                if changed { // this can be used regardless of 'err'.
                    any_changed = true
                    to_reregister[project] = true
                }
                if err != nil {
                    log.Printf("failed to ignore latest for %q; %v", k, err)
                }
                delete(to_reignore, k)
            }
            lock.Unlock()
            if any_changed {
                ch_reregister <- true
            }
        }
    }()

    // Listener to re-register projects.
    for {
        <- ch_reregister
        lock.Lock()
        for project, _ := range to_reregister {
            err := reregisterProject(rest_url, filepath.Join(registry, project))
            if err != nil {
                log.Printf("failed to reregister %q; %v", project, err)
            }
            delete(to_reregister, project)
        }
        lock.Unlock()
    }
}
