package main

import (
	"encoding/json"
	"flag"
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
	//JSONPayload := `{"NodeInfo":{"protocol_version":{"p2p":"8","block":"11","app":"0"},"id":"6f4d5f7ecb3d73ca291658d5d2368e23c34d9122","listen_addr":"tcp://0.0.0.0:26656","network":"theta-testnet-001","version":"0.34.27","channels":"40202122233038606100","moniker":"sr","other":{"tx_index":"on","rpc_address":"tcp://127.0.0.1:26657"}},"SyncInfo":{"latest_block_hash":"527D77A25A007353170611E0A3FB7163A281A47EF31DB9A46AAF25D42A5E1515","latest_app_hash":"48C6486DA403F98F13C22D7FF0114899CA88BFE2AC734B2A6C5E0C172046C8B6","latest_block_height":"16041033","latest_block_time":"2023-05-19T16:36:07.825229312Z","earliest_block_hash":"0C78904CC0CF2450C7A0845425FA2CA07A1003671B3DCD4CBAB9D160FFE56948","earliest_app_hash":"337522801C1F4279E170123DFA3D0DA604307F5F2D501DFE9B90507212F78D8E","earliest_block_height":"16040001","earliest_block_time":"2023-05-19T15:01:26.554070159Z","catching_up":true},"ValidatorInfo":{"Address":"AA20C839B1BBEF4BE431829A658C2EA6B28D1D22","PubKey":{"type":"tendermint/PubKeyEd25519","value":"gl8QDOmLJswTht/ID8IeTeXtYghwNUeysOglrSbjPEY="},"VotingPower":"0"}}`

	argument := "status"
	log.Printf("Executing %s %s", gaiadPath, argument)
	command, _ := exec.Command(gaiadPath, argument).Output()
	log.Printf("DEBUG: status JSON is: %v", command)
	return command
}

func handleMetrics() {
	var gaiaStatus GaiaStatus
	for {
		err := json.Unmarshal(GetGaiaStatus(), &gaiaStatus)
		if err != nil {
			log.Printf("Cannot unmarshal status JSON. Reason: %v", err)
			return
		}

		log.Printf("DEBUG: %v", gaiaStatus)

		if bn, err := strconv.ParseFloat(gaiaStatus.NodeInfo.ProtocolVersion.Block, 64); err == nil {
			log.Printf("DEBUG: block number set: %v", bn)
			blockNumber.Set(bn)
		}

		if pc, err := strconv.ParseFloat(gaiaStatus.NodeInfo.ProtocolVersion.P2P, 64); err == nil {
			log.Printf("DEBUG: peers count set: %v", pc)
			peersCount.Set(pc)
		}

		timeDelta := math.Round(time.Now().Sub(gaiaStatus.SyncInfo.LatestBlockTime).Seconds())
		log.Printf("DEBUG: time delta set: %v", timeDelta)
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
	log.Printf("Starting exporter on :%s%s for %s", listenPort, metricsPath, gaiadPath)

	go handleMetrics()

	err := metricsServe(":"+listenPort, metricsPath)
	if err != nil {
		log.Printf("Cannot serve metrics. Reason: %v", err)
		return
	}
}
