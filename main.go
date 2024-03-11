package main

import (
	"io"
	"log"
	"net/http"

	"github.com/alecthomas/kingpin/v2"
	kitlog "github.com/go-kit/log"
	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/hansmi/paperhooks/pkg/kpflag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

var webConfig = webflag.AddFlags(kingpin.CommandLine, ":8081")
var metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics").Default("/metrics").String()
var disableExporterMetrics = kingpin.Flag("web.disable-exporter-metrics", "Exclude metrics about the exporter itself").Bool()
var timeout = kingpin.Flag("scrape-timeout", "Maximum duration for a scrape").Default("1m").Duration()

func main() {
	var clientFlags client.Flags

	kpflag.RegisterClient(kingpin.CommandLine, &clientFlags)
	kingpin.Parse()

	client, err := clientFlags.Build()
	if err != nil {
		log.Fatal(err)
	}

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(newCollector(client, *timeout))

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

	logger := kitlog.NewLogfmtLogger(kitlog.StdlibWriter{})
	server := &http.Server{}

	if err := web.ListenAndServe(server, webConfig, logger); err != nil {
		log.Fatal(err)
	}
}
