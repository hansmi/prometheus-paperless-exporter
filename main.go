package main

import (
	"io"
	"log"
	"maps"
	"net/http"
	"slices"
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
var collectorsFlag = kingpin.Flag("collectors", "Comma-separated list of collector ids to enable. If empty all standard collectors are enabled.").String()

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

	// Parse collectors flag into a slice (comma separated)
	var enabledCollectors []string
	collectorsStr := strings.TrimSpace(*collectorsFlag)
	for _, s := range strings.Split(collectorsStr, ",") {
		if s = strings.TrimSpace(s); s != "" {
			enabledCollectors = append(enabledCollectors, s)
		}
	}

	// Validate collector ids
	if unknown := validateCollectorIDs(enabledCollectors); len(unknown) > 0 {
		// Build a helpful error message listing unknown and known ids
		knownKeys := slices.Collect(maps.Keys(knownCollectors()))
		// remote_version is special and can be enabled but is gated by enableRemoteNetwork
		knownKeys = append(knownKeys, remoteVersionCollectorID)
		log.Fatalf("unknown collector ids: %v. Known collector ids: %v", unknown, knownKeys)
	}

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(newCollector(client, *timeout, *enableRemoteNetwork, enabledCollectors))

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
