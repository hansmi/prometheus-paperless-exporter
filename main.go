package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/hansmi/paperhooks/pkg/kpflag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	promslogflag "github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

var webConfig = webflag.AddFlags(kingpin.CommandLine, ":8081")
var metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics").Default("/metrics").String()
var disableExporterMetrics = kingpin.Flag("web.disable-exporter-metrics", "Exclude metrics about the exporter itself").Bool()
var enableRemoteNetwork = kingpin.Flag("enable-remote-network", "Include calls to API endpoints that require public internet access for your paperless instance (e.g. checking for a paperless version)").Bool()
var timeout = kingpin.Flag("scrape-timeout", "Maximum duration for a scrape").Default("1m").Duration()
var remoteVersionInterval = kingpin.Flag("remote-version-interval", "Interval in seconds to poll remote version (unit: seconds)").Default("60").Int()

func main() {
	var clientFlags client.Flags

	promslogConfig := &promslog.Config{}
	promslogflag.AddFlags(kingpin.CommandLine, promslogConfig)

	kpflag.RegisterClient(kingpin.CommandLine, &clientFlags)
	kingpin.Parse()

	logger := promslog.New(promslogConfig)

	client, err := clientFlags.Build()
	if err != nil {
		log.Fatal(err)
	}

	reg := prometheus.NewPedanticRegistry()
	remoteVersionIntervalDuration := time.Duration(*remoteVersionInterval) * time.Second
	coll, stop := newCollector(client, *timeout, *enableRemoteNetwork, remoteVersionIntervalDuration)
	reg.MustRegister(coll)
	// Ensure background collectors are stopped on shutdown
	defer stop()

	if !*disableExporterMetrics {
		reg.MustRegister(
			collectors.NewBuildInfoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
			version.NewCollector("prometheus_paperless_exporter"),
		)
	}

	http.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `<html>
			<head><title>Paperless Exporter</title></head>
			<body>
			<h1>Paperless Exporter</h1>
			<p><a href="`+*metricsPath+`">Metrics</a></p>
			</body>
			</html>`)
	})

	server := &http.Server{}

	// Setup signal handling for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		// stop background collectors
		stop()

		// give the server some time to shutdown gracefully
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		os.Exit(0)
	}()

	if err := web.ListenAndServe(server, webConfig, logger); err != nil {
		log.Fatal(err)
	}
}
