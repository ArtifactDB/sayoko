package main

import (
    "os"
    "flag"
    "log"
    "time"
    "sync"
    "strings"
    "path/filepath"
)

func main() {
    gpath := flag.String("registry", "", "Path to the gobbler registry")
    surl := flag.String("url", "", "URL of the SewerRat instance")
    log_time := flag.Int("log", 10, "Interval in which to check for new logs, in minutes")
    full_time := flag.Int("full", 24, "Interval in which to do a full check, in hours")

    if *gpath == "" || *surl == "" {
        flag.Usage()
        os.Exit(1)
    }

    to_reignore := map[string]bool{}
    to_reregister := map[string]bool{}
    var lock sync.Mutex
    ch_reignore := make(chan bool)
    ch_reregister := make(chan bool)

    go func() {
        timer := time.NewTicker(time.Minute * time.Duration(*log_time))
        last_scan := time.Now()
        for {
            <-timer.C

            lock.Lock()
            new_last_scan, err := checkLogs(*gpath, *surl, to_reignore, to_reregister, last_scan)
            lock.Unlock()

            if err != nil {
                log.Printf("detected failures for log check; %v", err)
            }
            last_scan = new_last_scan // this can be set regardless of 'err'.
            ch_reignore <- true
        }
    }()

    go func() {
        timer := time.NewTicker(time.Hour * time.Duration(*full_time))
        for {
            <-timer.C

            lock.Lock()
            err := fullScan(*gpath, *surl, to_reignore)
            lock.Unlock()

            if err != nil {
                log.Printf("detected failures for log check; %v", err)
            }
            ch_reignore <- true
        }
    }()

    go func() {
        for {
            <- ch_reignore
            any_changed := false
            lock.Lock()

            for k, _ := range to_reignore {
                pos := strings.IndexByte(k, '/')
                project := k[:pos]
                asset := k[(pos + 1):]
                changed, err := ignoreNonLatest(*gpath, project, asset)

                if changed { // this can be used regardless of 'err'.
                    any_changed = true
                    to_reregister[project] = true
                }
                if err != nil {
                    log.Printf("failed to ignore latest for '%s/%s'; %v", project, asset, err)
                }
                delete(to_reignore, k)
            }

            lock.Unlock()
            if any_changed {
                ch_reregister <- true
            }
        }
    }()

    for {
        <- ch_reregister
        lock.Lock()
        for project, _ := range to_reregister {
            err := reregisterProject(*surl, filepath.Join(*gpath, project))
            if err != nil {
                log.Printf("failed to reregister %q; %v", project, err)
            }
            delete(to_reregister, project)
        }
        lock.Unlock()
    }
}
