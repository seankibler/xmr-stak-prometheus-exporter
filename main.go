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
  Results struct {
    Difficulty int `json:"diff_current"`
    SharesGood int `json:"shares_good"`
    SharesTotal int `json:"shares_total"`
    AvgResultTime float64 `json:"avg_time"`
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

func getXmrstakData() XmrstakApiResponse {
  resp, err := http.Get(xmrstakApiUrl)
  if err != nil {
    fmt.Printf("Error getting data from API URL \"%s\".\n", xmrstakApiUrl)
    os.Exit(1)
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)

  responseData := XmrstakApiResponse{}

  json.Unmarshal(body, &responseData)

	return responseData
}

func init() {
	// Metrics have to be registered to be exposed:
  if err := prometheus.Register(prometheus.NewGaugeFunc(
    prometheus.GaugeOpts{
      Name: "hashrate_total",
		  Help: "Total hashrate considering all devices",
    },
    func() float64 {
      xmrstakApiData := getXmrstakData()

      return xmrstakApiData.HashRate.Total[0]
    },
  )); err == nil {
    fmt.Println("Registered hashrate_total gauge")
  }

  if err := prometheus.Register(prometheus.NewGaugeFunc(
    prometheus.GaugeOpts{
      Name: "difficulty",
		  Help: "Hashing difficulty",
    },
    func() float64 {
      xmrstakApiData := getXmrstakData()

      return float64(xmrstakApiData.Results.Difficulty)
    },
  )); err == nil {
    fmt.Println("Registered difficulty gauge")
  }

  if err := prometheus.Register(prometheus.NewGaugeFunc(
    prometheus.GaugeOpts{
      Name: "avg_result_time",
		  Help: "Average time to submit a valid hash result",
    },
    func() float64 {
      xmrstakApiData := getXmrstakData()

      return float64(xmrstakApiData.Results.AvgResultTime)
    },
  )); err == nil {
    fmt.Println("Registered avg_result_time gauge")
  }

  if err := prometheus.Register(prometheus.NewGaugeFunc(
    prometheus.GaugeOpts{
      Name: "results_accepted",
		  Help: "Hash results accepted by mining pool",
    },
    func() float64 {
      xmrstakApiData := getXmrstakData()

      return float64(xmrstakApiData.Results.SharesGood)
    },
  )); err == nil {
    fmt.Println("Registered results_accepted gauge")
  }

  if err := prometheus.Register(prometheus.NewGaugeFunc(
    prometheus.GaugeOpts{
      Name: "results_total",
		  Help: "Hash results submitted to mining pool",
    },
    func() float64 {
      xmrstakApiData := getXmrstakData()

      return float64(xmrstakApiData.Results.SharesTotal)
    },
  )); err == nil {
    fmt.Println("Registered results_total gauge")
  }
}

func main() {
	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.
  fmt.Printf("Starting Prometheus metrics handler on %s:%s...", httpListenAddr, httpListenPort)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", httpListenAddr, httpListenPort), nil))
}

