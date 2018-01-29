package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"net/http"
	"log"
	"os"
	"github.com/btcsuite/btcd/btcjson"
	"encoding/json"
)

const namespace = "bitcoind"

func main() {
	log.Println("bitcoind-exporter starting...")
	defer log.Println("bitcoind-exporter stopping...")

	config := rpcclient.ConnConfig{
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	flagset := pflag.NewFlagSet("bitcoind-exporter", pflag.PanicOnError)
	flagset.StringVarP(&config.Host,"bitcoind-host", "h", "127.0.0.1:9332", "")
	flagset.StringVarP(&config.User,"bitcoind-user", "u", "bitcoind", "")
	flagset.StringVarP(&config.Pass,"bitcoind-password", "p", "password", "")
	listenTo := flagset.StringP("listen-to", "l", "0.0.0.0:8452", "")
	flagset.Parse(os.Args[1:])

	client, err := rpcclient.New(&config, nil)
	if err != nil {
		panic(err)
	}

	log.Printf("trying to connect to bitcoind at %q...\n", config.Host)
	_, err = client.GetInfo()
	if err != nil {
		panic(err)
	}
	log.Printf("successfuly connected to bitcoind\n")

	defer client.Shutdown()

	prometheus.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "blockchain",
		Name: "block_count",
		Help: "bitcoind's current block count",
	}, func() float64 {
		count, err := client.GetBlockCount()
		if err != nil {
			panic(err)
		}

		return float64(count)
	}), prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "blockchain",
		Name: "difficulty",
		Help: "bitcoind's current difficulty",
	}, func() float64 {
		difficulty, err := client.GetDifficulty()
		if err != nil {
			panic(err)
		}

		return difficulty
	}), prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "mempool",
		Name: "transaction_count",
		Help: "bitcoind's current transaction count in mempool",
	}, func() float64 {
		transactionsHashes, err := client.GetRawMempool()
		if err != nil {
			panic(err)
		}

		return float64(len(transactionsHashes))
	}), prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "network",
		Name: "connections_count",
		Help: "bitcoind's current connections count",
	}, func() float64 {
		rawResponse, err := client.RawRequest("getnetworkinfo", nil)
		if err != nil {
			panic(err)
		}
		result := &btcjson.GetNetworkInfoResult{}
		err = json.Unmarshal(rawResponse, result)
		if err != nil {
			panic(err)
		}

		return float64(result.Connections)
	}))

	log.Printf("starting to serve metrics on %q...\n", *listenTo)
	if err := http.ListenAndServe(*listenTo, promhttp.Handler()); err != nil {
		panic(err)
	}
}