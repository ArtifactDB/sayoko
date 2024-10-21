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

func isProjectRegistered(registry string, rest_url string, project string) (bool, error) {
    path := filepath.Join(registry, project)
    resp, err := http.Get(rest_url + "/registered?contains_path=" + url.QueryEscape(path))
    if err != nil {
        return false, fmt.Errorf("failed to check if %q is registered; %w", project, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        err := parseFailure(resp)
        return false, fmt.Errorf("failed to check if %q is registered; %w", project, err)
    }

    dec := json.NewDecoder(resp.Body)
    output := []interface{}{}
    err = dec.Decode(&output)
    if err != nil {
        return false, fmt.Errorf("failed to parse /registered response; %w", err)
    }

    return len(output) > 0, nil
}

func isProjectRegisteredWithCache(registry string, url string, project string, cache map[string]bool) (bool, error) {
    is_reg, ok := cache[project]
    if ok {
        return is_reg, nil
    }

    is_reg, err := isProjectRegistered(registry, url, project)
    if err != nil {
        return is_reg, err
    }

    cache[project] = is_reg
    return is_reg, err
}

func reregisterProject(rest_url string, project_dir string) error {
    {
        b, err := json.Marshal(map[string]string{ "path": project_dir })
        if err != nil {
            return fmt.Errorf("failed to create initialization request body for %q; %w", project_dir, err)
        }

        r := bytes.NewReader(b)
        resp, err := http.Post(rest_url + "/register/start", "application/json", r)
        if err != nil {
            return fmt.Errorf("failed to initialize registration for %q; %w", project_dir, err)
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 300 {
            err := parseFailure(resp)
            return fmt.Errorf("failed to initialize registration for %q; %w", project_dir, err)
        }

        decoded := struct {
            Code string `json:"code"`
        }{}
        dec := json.NewDecoder(resp.Body)
        err = dec.Decode(&decoded)
        if err != nil {
            return fmt.Errorf("failed to parse initialization response for %q; %w", project_dir, err) 
        }

        code_path := filepath.Join(project_dir, decoded.Code)
        handle, err := os.OpenFile(code_path, os.O_CREATE, 0644)
        if err != nil {
            return fmt.Errorf("failed to create registration code in %q; %w", project_dir, err)
        }
        handle.Close()
        defer os.Remove(code_path)
    }

    {
        b, err := json.Marshal(map[string]string{ "path": project_dir })
        if err != nil {
            return fmt.Errorf("failed to create registration completion request body for %q; %w", project_dir, err)
        }

        r := bytes.NewReader(b)
        resp, err := http.Post(rest_url + "/register/finish", "application/json", r)
        if err != nil {
            return fmt.Errorf("failed to finish registration for %q; %w", project_dir, err)
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 300 {
            err := parseFailure(resp)
            return fmt.Errorf("failed to finish registration for %q; %w", project_dir, err)
        }
    }

    return nil
}
