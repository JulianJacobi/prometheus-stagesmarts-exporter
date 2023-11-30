package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "io"
	"log"
	"net/http"
    "net/url"

    prom "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    address = flag.String("listen-address", "127.0.0.1", "Address the webserver listen on.")
    port = flag.Int("port", 9005, "Port the webserver binds to.")
)

func main() {
    flag.Parse()
    
    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        target, ok := r.URL.Query()["target"]
        w.Header().Add("ContentType", "text")
        if !ok {
            w.WriteHeader(400)
            w.Write([]byte("Missing parameter 'target'!"))
            return
        }
        resp, err := http.Get(fmt.Sprintf("http://%s:8080/api/getcurrentpduvalues", target))
        if err != nil {
            urlErr := err.(*url.Error)
            if urlErr.Timeout() {
                // handle down target
                return
            } else {
                log.Printf("Error while requesting PDU: %s", err)
                w.WriteHeader(500)
                w.Write([]byte("Internal Server Error"))
                return
            }
        }
        defer resp.Body.Close()
        rawData, err := io.ReadAll(resp.Body)
        if err != nil {
            log.Printf("Error while reading response body: %s", err)
            w.WriteHeader(500)
            w.Write([]byte("Internal Server Error"))
            return
        }
        var value PduValueResponse
        err = json.Unmarshal(rawData, &value)
        if err != nil {
            log.Printf("Error while parsing response JSON: %s", err)
            w.WriteHeader(500)
            w.Write([]byte("Internal Server Error"))
            return
        }
        reg := prom.NewRegistry()
        for _, v := range value.GetMetrics() {
            reg.MustRegister(v)
        }
        promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
    })

    listenAddress := fmt.Sprintf("%s:%d", *address, *port)

    log.Printf("Start listening on %s", listenAddress)
    log.Fatal(http.ListenAndServe(listenAddress, nil))
}
