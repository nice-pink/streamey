package main

import (
	"flag"
	"strings"
	"sync"
	"time"

	"github.com/nice-pink/audio-tool/pkg/audio/encodings"
	"github.com/nice-pink/audio-tool/pkg/network"
	"github.com/nice-pink/audio-tool/pkg/util"
	"github.com/nice-pink/goutil/pkg/filesystem"
	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/configmanager"
	"github.com/nice-pink/streamey/pkg/metricmanager"
	"github.com/nice-pink/streamey/pkg/miniomanager"
)

var wg sync.WaitGroup

const (
	delay       int64  = 3600
	bucket      string = "data"
	minioFolder string = "audio"
)

func main() {
	log.Info("--- Start readey ---")

	// flags
	url := flag.String("url", "", "Stream url")
	timeout := flag.Int("timeout", 30, "Timeout. Default: 30sec")
	validate := flag.String("validate", "", "Validation type. [audio, privateBit]")
	outputFilepath := flag.String("outputFilepath", "", "[Optional] Output file path, if data should be dumped to file.")
	reconnect := flag.Bool("reconnect", false, "[Optional] Reconnect on any interruption.")
	minioConfig := flag.String("minioConfig", "", "[Optional] Json config file for minio. Use minio if defined.")
	minioCleanUpAfterSec := flag.Int64("minioCleanUpAfterSec", 0, "[Optional] Cleanup minio bucket after seconds.")
	config := flag.String("config", "", "Config file.")
	verbose := flag.Bool("verbose", false, "Verbose Logging.")
	metrics := flag.Bool("metrics", false, "Add metrics.")
	metricPrefix := flag.String("metricPrefix", "streamey_", "Metric prefix.")
	metricPort := flag.Int("metricPort", 9090, "Metric port.")
	flag.Parse()

	var c configmanager.ReadConfig
	if *config != "" {
		c = configmanager.GetReadConfig(*config)
	}

	// start metrics server
	metricsControl := util.MetricsControl{Enabled: false}
	if *metrics {
		metricsControl.Prefix = *metricPrefix
		metricsControl.Labels = map[string]string{"url": *url}
		go metricmanager.Listen(*metricPort)
	}

	// read stream
	go ReadStream(*url, *outputFilepath, *reconnect, false, *timeout, c, *validate, metricsControl, *verbose)

	// start minio sync
	goRoutineCounter := 1
	if *minioConfig != "" {
		goRoutineCounter++

		go ManageMinio(c, delay, *outputFilepath, *minioCleanUpAfterSec, *reconnect)
	}

	// wait for go routines to be done
	wg.Add(goRoutineCounter)
	wg.Wait()
}

// stream

func ReadStream(url string, outputFilepath string, reconnect, failEarly bool, timeout int, config configmanager.ReadConfig, validate string, metricsControl util.MetricsControl, verbose bool) {
	connection := network.NewConnection(url, "", 80, 0, time.Duration(timeout), network.HttpConnection, metricsControl)
	var validator network.DataValidator
	if strings.ToLower(validate) == "audio" {
		log.Newline()
		log.Info("### Audio validation")
		// expectations := audio.Expectations{
		// 	IsCBR: true,
		// 	Encoding: audio.Encoding{
		// 		Bitrate:  256,
		// 		IsStereo: true,
		// 	},
		// }
		expectations := config.Expectations
		expectations.Print()
		log.Info("###")
		log.Newline()
		validator = encodings.NewEncodingValidator(true, failEarly, expectations, metricsControl, verbose)
	} else if strings.ToLower(validate) == "privatebit" {
		validator = encodings.NewPrivateBitValidator(true, encodings.GuessAudioType(url), metricsControl, verbose)
	} else {
		validator = network.DummyValidator{}
	}
	connection.ReadStream(outputFilepath, reconnect, validator)
	wg.Done()
}

// minio

func ManageMinio(config configmanager.ReadConfig, delay int64, localFolder string, minioCleanUpAfterSec int64, loop bool) {
	useSsl := true
	m := miniomanager.NewMinioManager(config.Minio, useSsl)

	// run loop
	duration := time.Duration(delay) * time.Second
	for {
		// start with the dalay
		time.Sleep(duration)

		// push
		PushFiles(m, localFolder)

		// delete
		//day := int64(3600 * 24)
		m.DeleteFiles(bucket, minioFolder, minioCleanUpAfterSec)

		if !loop {
			break
		}
	}
	wg.Done()
}

func PushFiles(m *miniomanager.MinioManger, folder string) {
	files := filesystem.ListFiles(folder, 0, true)

	for _, file := range files {
		m.PutFile(bucket, minioFolder, folder, file, true)
	}
}
