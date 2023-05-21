package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"math"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

var (
	listenPort,
	metricsPath,
	gaiadPath = ParseFlags()

	blockNumber = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "gaia_block_number",
			Help: "Block number",
		})

	peersCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "gaia_peers_count",
			Help: "Peers count",
		})

	blockTimeDelta = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "gaia_block_time_delta_seconds",
			Help: "Block sync delta seconds",
		})
)

type GaiaStatus struct {
	NodeInfo struct {
		ProtocolVersion struct {
			P2P   string `json:"p2p"`
			Block string `json:"block"`
			App   string `json:"app"`
		} `json:"protocol_version"`
		Id         string `json:"id"`
		ListenAddr string `json:"listen_addr"`
		Network    string `json:"network"`
		Version    string `json:"version"`
		Channels   string `json:"channels"`
		Moniker    string `json:"moniker"`
		Other      struct {
			TxIndex    string `json:"tx_index"`
			RpcAddress string `json:"rpc_address"`
		} `json:"other"`
	} `json:"NodeInfo"`
	SyncInfo struct {
		LatestBlockHash     string    `json:"latest_block_hash"`
		LatestAppHash       string    `json:"latest_app_hash"`
		LatestBlockHeight   string    `json:"latest_block_height"`
		LatestBlockTime     time.Time `json:"latest_block_time"`
		EarliestBlockHash   string    `json:"earliest_block_hash"`
		EarliestAppHash     string    `json:"earliest_app_hash"`
		EarliestBlockHeight string    `json:"earliest_block_height"`
		EarliestBlockTime   time.Time `json:"earliest_block_time"`
		CatchingUp          bool      `json:"catching_up"`
	} `json:"SyncInfo"`
	ValidatorInfo struct {
		Address string `json:"Address"`
		PubKey  struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"PubKey"`
		VotingPower string `json:"VotingPower"`
	} `json:"ValidatorInfo"`
}

func ParseFlags() (string, string, string) {
	listenPort := flag.String("port", "19001", "Port to listen metrics requests")
	metricsPath := flag.String("metrics-path", "/metrics", "Path to expose metrics.")
	gaiadPath := flag.String("gaiad-bin-path", "/usr/sbin/gaiad", "Path to gaiad binary.")

	flag.Parse()

	return *listenPort, *metricsPath, *gaiadPath
}

func GetGaiaStatus() []byte {
	arg := "status"
	cmd := exec.Command(gaiadPath, arg)

	//var status bytes.Buffer
	//cmd.Stdout = &status

	log.Printf("Executing %s %s", gaiadPath, arg)
	//if err := cmd.Run(); err != nil {
	//	log.Fatal(err)
	//}
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	fmt.Print(string(out))

	return out
}

func handleMetrics() {
	var gaiaStatus GaiaStatus
	for {
		err := json.Unmarshal(GetGaiaStatus(), &gaiaStatus)
		if err != nil {
			log.Printf("Cannot unmarshal status JSON. Reason: %v\n", err)
			return
		}

		log.Printf("DEBUG: %v\n", gaiaStatus)

		if bn, err := strconv.ParseFloat(gaiaStatus.NodeInfo.ProtocolVersion.Block, 64); err == nil {
			log.Printf("DEBUG: block number set: %v", bn)
			blockNumber.Set(bn)
		}

		if pc, err := strconv.ParseFloat(gaiaStatus.NodeInfo.ProtocolVersion.P2P, 64); err == nil {
			log.Printf("DEBUG: peers count set: %v", pc)
			peersCount.Set(pc)
		}

		timeDelta := math.Round(time.Now().Sub(gaiaStatus.SyncInfo.LatestBlockTime).Seconds())
		log.Printf("DEBUG: time delta set: %v\n", timeDelta)
		blockTimeDelta.Set(timeDelta)

		time.Sleep(10 * time.Second)
	}
}

func metricsServe(listenAddress, metricsPath string) error {
	http.Handle(metricsPath, promhttp.Handler())
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`
               <html>
               <head><title>Gaiad metrics exporter</title></head>
               <body>
               <h1>Gaiad metrics exporter</h1>
               <p><a href='` + metricsPath + `'>Metrics</a></p>
               </body>
               </html>
            `))
		})
	return http.ListenAndServe(listenAddress, nil)
}

func main() {
	log.Printf("Starting exporter on :%s%s for %s\n", listenPort, metricsPath, gaiadPath)

	go handleMetrics()

	err := metricsServe(":"+listenPort, metricsPath)
	if err != nil {
		log.Printf("Cannot serve metrics. Reason: %v\n", err)
		return
	}
}
