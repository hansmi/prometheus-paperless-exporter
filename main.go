package main

import (
	"io"
	"log"
	"net/http"
	"strings"

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
var collectorsFlag = kingpin.Flag("collectors", "Comma-separated list of collectors to enable. If empty all standard collectors are enabled.").String()

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

	var enabledCollectors []string

	// Parse comma-separated collectors flag into a slice.
	for _, s := range strings.Split(strings.TrimSpace(*collectorsFlag), ",") {
		if s = strings.TrimSpace(s); s != "" {
			enabledCollectors = append(enabledCollectors, s)
		}
	}

	collector, err := newCollector(client, *timeout, *enableRemoteNetwork, enabledCollectors)
	if err != nil {
		log.Fatalf("Collector: %v", err)
	}

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)

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

	if err := web.ListenAndServe(server, webConfig, logger); err != nil {
		log.Fatal(err)
	}
}
