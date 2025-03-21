package main

import (
    "os"
    "flag"
    "log"
    "fmt"
    "time"
    "sync"
    "path/filepath"
    "errors"
)

func retrieveLastScanTime(last_scan_path string) time.Time {
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

func depositLastScanTime(last_scan time.Time, last_scan_path string) {
    err := os.WriteFile(last_scan_path, []byte(last_scan.Format(time.RFC3339)), 0644)
    if err != nil {
        log.Printf("failed to write the last scan time; %v", err)
    }
}

func main() {
    gpath := flag.String("registry", "", "Path to the gobbler registry")
    surl := flag.String("url", "", "URL of the SewerRat instance")
    log_time := flag.Int("log", 10, "Interval in which to check for new logs, in minutes")
    full_time := flag.Int("full", 168, "Interval in which to do a full check, in hours")
    tpath := flag.String("timestamp", ".sayoko_last_scan", "Path to the last scan timestamp")
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

    var lock sync.Mutex

    // Timer to inspect logs.
    go func() {
        last_scan_path := *tpath
        last_scan := retrieveLastScanTime(last_scan_path)
        timer := time.NewTicker(time.Minute * time.Duration(*log_time))
        for {
            lock.Lock()
            new_last_scan, err := processLogs(rest_url, registry, last_scan)
            lock.Unlock()
            if err != nil {
                log.Printf("detected failures for log check; %v", err)
            }
            if last_scan != new_last_scan { // new_last_scan can be used regardless of 'err'.
                last_scan = new_last_scan
                depositLastScanTime(last_scan, last_scan_path)
            }
            <-timer.C
        }
    }()

    // Timer to scan the entire registry.
    timer := time.NewTicker(time.Hour * time.Duration(*full_time))
    for {
        lock.Lock()
        err := fullScan(rest_url, registry)
        lock.Unlock()
        if err != nil {
            log.Printf("detected failures for log check; %v", err)
        }
        <-timer.C
    }
}
