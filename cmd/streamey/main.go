package main

import (
	"flag"
	"sync"

	"github.com/nice-pink/audio-tool/pkg/util"
	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/configmanager"
	"github.com/nice-pink/streamey/pkg/metricmanager"
	"github.com/nice-pink/streamey/pkg/streamer"
)

var wg sync.WaitGroup

func main() {
	log.Info("--- Start streamey ---")

	// flags
	url := flag.String("url", "", "Destination url")
	// filepath := flag.String("filepath", "", "File to stream.")
	// reconnect := flag.Bool("reconnect", false, "[Optional] Reconnect on any interruption.")
	// bitrate := flag.Int64("bitrate", 0, "Send bitrate.")
	// sr := flag.Int("sr", 44100, "Sampe rate.")
	// metaUrl := flag.String("metaUrl", "", "Metadata sink url.")
	// metaBody := flag.String("metaBody", "", "Metadata body string or file (start with @).")
	// test := flag.Bool("test", false, "Test sending with internal reader.")
	verbose := flag.Bool("verbose", false, "Verbose logging.")
	// isIcecast := flag.Bool("icecast", false, "Send icecast.")
	// validateTest := flag.String("validateTest", "", "Validation test.")
	metrics := flag.Bool("metrics", false, "Add metrics.")
	metricPrefix := flag.String("metricPrefix", "streamey_", "Metric prefix.")
	metricPort := flag.Int("metricPort", 9090, "Metric port.")
	configFilepath := flag.String("config", "", "Config filepath")
	flag.Parse()

	// start metrics server
	metricsControl := util.MetricsControl{Enabled: false}
	if *metrics {
		metricsControl.Prefix = *metricPrefix
		metricsControl.Labels = map[string]string{"url": *url}
		go metricmanager.Listen(*metricPort)
	}

	config := configmanager.GetStreamConfig(*configFilepath)

	// streamUrl := *url
	// if *test {
	// 	goRoutineCounter++
	// 	go Receive(*validateTest, metricsControl, *verbose)
	// 	// overwrite url
	// 	streamUrl = "localhost:9999"
	// }

	for _, item := range config.Items {
		go streamer.Stream(item, metricsControl, &wg, *verbose)
		wg.Add(1)
	}

	wg.Wait()
}
