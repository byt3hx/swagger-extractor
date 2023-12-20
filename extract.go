package main

import (
    "encoding/json"
    "html/template"
    "io/ioutil"
    "net/http"
    "net/url"
    "strings"
    "fmt"
    "crypto/tls"
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

)

type SwaggerSpec struct {
    Servers []Server `json:"servers"`
    Paths   map[string]map[string]map[string]interface{} `json:"paths"`
}


type Server struct {
    URL string `json:"url"`
}

type APIRequestDetails struct {
    SwaggerURL string
    BaseURL    string
    ProxyURL   string
    Headers    string 
    Results    []string
}


func extractUrlsAndParamsFromSwagger(swaggerSpec SwaggerSpec, defaultBaseUrl string) []map[string]interface{} {
    var apiDetails []map[string]interface{}

    baseUrl := defaultBaseUrl
    if len(swaggerSpec.Servers) > 0 && swaggerSpec.Servers[0].URL != "" {
        baseUrl = swaggerSpec.Servers[0].URL
    }

    for path, methods := range swaggerSpec.Paths {
        for method, details := range methods {
            if strings.ToUpper(method) != "DELETE" {
                fullPath := baseUrl + path
                params, _ := details["parameters"].([]interface{})
                apiDetail := map[string]interface{}{
                    "method": method,
                    "url":    fullPath,
                    "params": params,
                }
                apiDetails = append(apiDetails, apiDetail)
            }
        }
    }

    return apiDetails
}

func makeRequests(apiDetails []map[string]interface{}, proxyUrl string, headersStr string) []string {
    headers := make(http.Header)
    for _, line := range strings.Split(headersStr, "\n") {
        parts := strings.SplitN(line, ":", 2)
        if len(parts) == 2 {
            headers.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
        }
    }
    var results []string
    var client *http.Client

    if proxyUrl != "" {
        proxy, err := url.Parse(proxyUrl)
        if err != nil {
            return []string{fmt.Sprintf("Invalid proxy URL: %s", err.Error())}
        }
        transport := &http.Transport{Proxy: http.ProxyURL(proxy)}
        client = &http.Client{Transport: transport}
    } else {
        client = &http.Client{}
    }

    for _, detail := range apiDetails {
        urlStr := detail["url"].(string)
        method := strings.ToUpper(detail["method"].(string))
        fmt.Printf("Making %s request to %s\n", method, urlStr) 
        var req *http.Request
        var err error

        switch method {
        case "GET":
            req, err = http.NewRequest(http.MethodGet, urlStr, nil)
        case "POST":
            payload := strings.NewReader("{}") 
            req, err = http.NewRequest(http.MethodPost, urlStr, payload)
            req.Header.Set("Content-Type", "application/json")
        case "PATCH":
            payload := strings.NewReader("{}")
            req, err = http.NewRequest(http.MethodPatch, urlStr, payload)
            req.Header.Set("Content-Type", "application/json")
        case "PUT":
            payload := strings.NewReader("{}") 
            req, err = http.NewRequest(http.MethodPut, urlStr, payload)
            req.Header.Set("Content-Type", "application/json")
        default:
            results = append(results, fmt.Sprintf("Unsupported method: %s", method))
            continue
        }

        if err != nil {
            results = append(results, fmt.Sprintf("Error creating request: %s", err.Error()))
            continue
        }
        for key, values := range headers {
            for _, value := range values {
                req.Header.Add(key, value)
            }
        }
        resp, err := client.Do(req)
        if err != nil {
            results = append(results, fmt.Sprintf("Request to %s failed: %s", urlStr, err.Error()))
            continue
        }

        results = append(results, fmt.Sprintf("Response For : %s %d", urlStr , resp.StatusCode))
        resp.Body.Close()
    }
    return results
}


func handleForm(w http.ResponseWriter, r *http.Request) {
    tpl := template.Must(template.ParseFiles("form.html"))

    if r.Method != http.MethodPost {
        tpl.Execute(w, nil)
        return
    }

    details := APIRequestDetails{
        SwaggerURL: r.FormValue("swaggerurl"),
        BaseURL:    r.FormValue("baseurl"),
        ProxyURL:   r.FormValue("proxyurl"),
    }

    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: tr}

    resp, err := client.Get(details.SwaggerURL)
    if err != nil {
        http.Error(w, "Failed to fetch Swagger JSON: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        http.Error(w, "Failed to read response body: "+err.Error(), http.StatusInternalServerError)
        return
    }

    var swaggerSpec SwaggerSpec
    err = json.Unmarshal(body, &swaggerSpec)
    if err != nil {
        http.Error(w, "Failed to unmarshal Swagger JSON: "+err.Error(), http.StatusInternalServerError)
        return
    }

    apiDetails := extractUrlsAndParamsFromSwagger(swaggerSpec, details.BaseURL)
    details.Headers = r.FormValue("headers")
    details.Results = makeRequests(apiDetails, details.ProxyURL, details.Headers)

    tpl.Execute(w, details)
}

func main() {
    stopChan := make(chan os.Signal, 1)
    signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

    http.HandleFunc("/", handleForm)
    srv := &http.Server{Addr: ":8081", Handler: nil}
    go func() {
        log.Println("Starting server... on http://localhost:8081")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("ListenAndServe(): %v", err)
        }
    }()

    <-stopChan
    log.Println("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server Shutdown: %v", err)
    }
    log.Println("Server gracefully stopped")
}
