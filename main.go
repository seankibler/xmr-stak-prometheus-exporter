package main

import (
	"log"
	"net/http"
  "os"
  "fmt"
  "io/ioutil"
  "encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type XmrstakApiResponse struct {
  HashRate struct {
    Total []float64
  }
}

const (
  defaultHttpAddr = ""
  defaultHttpPort = "8080"
  defaultXmrstakApiUrl = "http://localhost:10000/api.json"
)

var (
  httpListenAddr = getEnv("HTTP_LISTEN_ADDR", defaultHttpAddr)
  httpListenPort = getEnv("HTTP_LISTEN_PORT", defaultHttpPort)
  xmrstakApiUrl  = getEnv("XMRSTAK_API_URL",  defaultXmrstakApiUrl)
)

func getEnv(key, fallback string) string {
  if value, ok := os.LookupEnv(key); ok {
    return value
  }
  return fallback
}

func getHashrate() float64 {
  resp, err := http.Get(xmrstakApiUrl)
  if err != nil {
    fmt.Printf("Error getting data from API URL \"%s\".\n", xmrstakApiUrl)
    os.Exit(1)
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)

  responseData := XmrstakApiResponse{}

  json.Unmarshal(body, &responseData)

	return responseData.HashRate.Total[0]
}

func init() {
	// Metrics have to be registered to be exposed:
  if err := prometheus.Register(prometheus.NewGaugeFunc(
    prometheus.GaugeOpts{
      Name: "hashrate_total",
		  Help: "Total hashrate considering all devices",
    },
    func() float64 { return getHashrate() },
  )); err == nil {
    fmt.Println("Registered hashrate_total gauge")
  }
}

func main() {
	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
  fmt.Printf("Starting Prometheus metrics handler on %s:%s...", httpListenAddr, httpListenPort)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", httpListenAddr, httpListenPort), nil))
}

