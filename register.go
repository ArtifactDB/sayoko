package main

import (
    "io"
    "errors"
    "net/http"
    "net/url"
    "path/filepath"
    "encoding/json"
    "fmt"
    "bytes"
    "os"
)

type errorResponse struct {
    Reason *string `json:"reason"`
}

func parseFailure(resp *http.Response) error {
    ct := resp.Header.Get("Content-Type") 
    if ct == "application/json" {
        dec := json.NewDecoder(resp.Body)
        errinfo := errorResponse{}
        err := dec.Decode(&errinfo)
        if err != nil {
            return fmt.Errorf("failed to parse error response (%q); %w", resp.StatusCode, err)
        }

        if errinfo.Reason == nil {
            return fmt.Errorf("lack of 'reason' in the error response (%q); %w", resp.StatusCode, err)
        }

        return errors.New(*(errinfo.Reason))
    }

    if ct == "text/plain" {
        b, err := io.ReadAll(resp.Body)
        if err != nil {
            return fmt.Errorf("failed to parse error response (%q); %w", resp.StatusCode, err)
        }
        return errors.New(string(b))
    }

    return fmt.Errorf("unknown content type %q for error response (%q)", ct, resp.StatusCode)
}

type registeredDirectory struct {
    Path string `json:"path"`
}

func listRegisteredDirectoriesRaw(url string) ([]registeredDirectory, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        err := parseFailure(resp)
        return nil, err
    }

    dec := json.NewDecoder(resp.Body)
    output := []registeredDirectory{}
    err = dec.Decode(&output)
    if err != nil {
        return nil, err
    }

    return output, err
}

func listRegisteredSubdirectories(rest_url, dir string) ([]string, error) {
    output, err := listRegisteredDirectoriesRaw(rest_url + "/registered?within_path=" + url.QueryEscape(dir))
    if err != nil {
        return nil, fmt.Errorf("failed to list subdirectories of %q; %w", dir, err)
    }
    collected := []string{}
    for _, val := range output {
        rel, err := filepath.Rel(dir, val.Path)
        if err == nil && filepath.IsLocal(rel) {
            collected = append(collected, rel)
        }
    }
    return collected, nil
}

func registerDirectoryRaw(rest_url, dir string, names []string, register bool) error {
    endpt := "register"
    msg := "registration"
    if !register {
        endpt = "deregister"
        msg = "deregistration"
    }

    {
        payload := map[string]interface{}{ "path": dir }
        b, err := json.Marshal(payload)
        if err != nil {
            return fmt.Errorf("failed to create initialization request body for %q; %w", dir, err)
        }

        r := bytes.NewReader(b)
        resp, err := http.Post(rest_url + "/" + endpt + "/start", "application/json", r)
        if err != nil {
            return fmt.Errorf("failed to initialize %s for %q; %w", msg, dir, err)
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 300 {
            err := parseFailure(resp)
            return fmt.Errorf("failed to initialize %s for %q; %w", msg, dir, err)
        }

        decoded := struct {
            Code string `json:"code"`
            Status string `json:"status"`
        }{}
        dec := json.NewDecoder(resp.Body)
        err = dec.Decode(&decoded)
        if err != nil {
            return fmt.Errorf("failed to parse initialization response for %q; %w", dir, err) 
        }

        if !register && decoded.Status == "SUCCESS" {
            return nil
        }

        code_path := filepath.Join(dir, decoded.Code)
        handle, err := os.OpenFile(code_path, os.O_CREATE, 0644)
        if err != nil {
            return fmt.Errorf("failed to create %s code in %q; %w", msg, dir, err)
        }
        handle.Close()
        defer os.Remove(code_path)
    }

    {
        payload := map[string]interface{}{ "path": dir }
        if register && names != nil{
            payload["base"] = names
        }
        b, err := json.Marshal(payload)
        if err != nil {
            return fmt.Errorf("failed to create registration completion request body for %q; %w", dir, err)
        }

        r := bytes.NewReader(b)
        resp, err := http.Post(rest_url + "/" + endpt + "/finish", "application/json", r)
        if err != nil {
            return fmt.Errorf("failed to finish %s for %q; %w", msg, dir, err)
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 300 {
            err := parseFailure(resp)
            return fmt.Errorf("failed to finish %s for %q; %w", msg, dir, err)
        }
    }

    return nil
}

func registerDirectory(rest_url, dir string, names []string) error {
    return registerDirectoryRaw(rest_url, dir, names, true)
}

func deregisterDirectory(rest_url, dir string) error {
    return registerDirectoryRaw(rest_url, dir, nil, false)
}

func deregisterSubdirectoriesRaw(rest_url, dir string, not_exists bool) error {
    url := rest_url + "/registered?within_path=" + url.QueryEscape(dir)
    if not_exists {
        url += "&exists=false"
    }
    output, err := listRegisteredDirectoriesRaw(url)
    if err != nil {
        return fmt.Errorf("failed to list subdirectories of %q; %w", dir, err)
    }
    all_errors := []error{}
    for _, val := range output {
        err := deregisterDirectory(rest_url, val.Path)
        all_errors = append(all_errors, err)
    }
    return errors.Join(all_errors...)
}

func deregisterAllSubdirectories(rest_url, dir string) error {
    return deregisterSubdirectoriesRaw(rest_url, dir, false)
}

func deregisterMissingSubdirectories(rest_url, dir string) error {
    return deregisterSubdirectoriesRaw(rest_url, dir, true)
}
