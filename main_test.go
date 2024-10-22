package main

import (
    "time"
    "testing"
)

func TestLastScanTime(t *testing.T) {
    last_scan := time.Now()
    const last_scan_path = ".sayoko_last_scan"
    depositLastScanTime(last_scan, last_scan_path)
    retrieved := retrieveLastScanTime(last_scan_path)
    if last_scan.Sub(retrieved).Abs() > time.Second { // needs some tolerance due to rounding of the stringified time.
        t.Fatalf("incorrect time value after a roundtrip (%v vs %v)", last_scan, retrieved)
    }
}

func TestUnpackKey(t *testing.T) {
    project, asset := unpackKey("foo/bar")
    if project != "foo" || asset != "bar" {
        t.Fatalf("unexpected project/asset split")
    }
}
