package main

import (
	"errors"
	"flag"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nice-pink/audio-tool/pkg/network"
	"github.com/nice-pink/audio-tool/pkg/util"
	"github.com/nice-pink/goutil/pkg/data"
	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/icecast"
	"github.com/nice-pink/streamey/pkg/metricmanager"
)

const (
	HTTP_VERSION string = "1.1"
	CHUNK_SIZE   int    = 1024
)

var wg sync.WaitGroup

func main() {
	log.Info("--- Start streamey ---")

	// flags
	url := flag.String("url", "", "Destination url")
	filepath := flag.String("filepath", "", "File to stream.")
	reconnect := flag.Bool("reconnect", false, "[Optional] Reconnect on any interruption.")
	bitrate := flag.Int64("bitrate", 0, "Send bitrate.")
	sr := flag.Int("sr", 44100, "Sampe rate.")
	metaUrl := flag.String("metaUrl", "", "Metadata sink url.")
	metaBody := flag.String("metaBody", "", "Metadata body string or file (start with @).")
	test := flag.Bool("test", false, "Test sending with internal reader.")
	verbose := flag.Bool("verbose", false, "Verbose logging.")
	isIcecast := flag.Bool("icecast", false, "Send icecast.")
	validateTest := flag.String("validateTest", "", "Validation test.")
	metrics := flag.Bool("metrics", false, "Add metrics.")
	metricPrefix := flag.String("metricPrefix", "streamey_", "Metric prefix.")
	metricPort := flag.Int("metricPort", 9090, "Metric port.")
	flag.Parse()

	// read file
	data := GetData(*filepath)
	if len(data) == 0 {
		log.Error("no data in file", filepath)
		os.Exit(2)
	}

	// start metrics server
	metricsControl := util.MetricsControl{Enabled: false}
	if *metrics {
		metricsControl.Prefix = *metricPrefix
		metricsControl.Labels = map[string]string{"url": *url}
		go metricmanager.Listen(*metricPort)
	}

	goRoutineCounter := 0

	streamUrl := *url
	if *test {
		goRoutineCounter++
		go Receive(*validateTest, metricsControl, *verbose)
		// overwrite url
		streamUrl = "localhost:9999"
	}

	goRoutineCounter++
	go Stream(streamUrl, float64(*bitrate), *sr, data, *reconnect, *isIcecast, *metaUrl, *metaBody, metricsControl, *verbose)

	wg.Add(goRoutineCounter)
	wg.Wait()
}

func GetData(filepath string) []byte {
	log.Info("Get data from", filepath)
	filepathFinal := filepath
	if strings.HasPrefix(filepath, "http") {
		filepathFinal = "file.mp3"
		util.DownloadFile(filepath, filepathFinal)
	}

	// is local file
	file, err := os.Open(filepathFinal)
	if err != nil {
		log.Err(err, "Cannot open file.")
	}
	data, err := io.ReadAll(file)
	if err != nil {
		log.Err(err, "Cannot read file.")
	}
	return data
}

func Stream(url string, bitrate float64, sampleRate int, data []byte, reconnect bool, isIcecast bool, metaUrl, metaBody string, metrics util.MetricsControl, verbose bool) {
	log.Info("Stream data with bitrate", bitrate, "to", url)
	var icyAdd icecast.IcyAdd
	var err error
	if isIcecast {
		// overwrite url with icy address
		icyAdd, err = icecast.GetIcyAddress(url)
		if err != nil {
			log.Error("Icy address")
			os.Exit(2)
		}
		url = icyAdd.Domain + icyAdd.Port
	}

	// stream
	log.Info("Conn to url", url)
	connection := network.NewConnection(url, "", 80, 0, time.Duration(30), network.HttpConnection, metrics)
	socketConn, err := connection.GetSocketConn()
	if err != nil {
		log.Err(err, "cannot establish socket connection to", url)
		os.Exit(2)
	}
	defer socketConn.Close()

	// send metadata
	httpClient := http.Client{}
	metaRequest := GetMetaRequest(metaUrl, metaBody)
	metaSendFn := func() error {
		if metaRequest == nil {
			return nil
		}
		resp, err := httpClient.Do(metaRequest)
		if err != nil {
			log.Err(err, "send metadata error")
		}
		if resp.StatusCode != 200 {
			log.Error("status code != 200:", resp.StatusCode)
		}
		return err
	}

	// init function
	if isIcecast {
		log.Info("Establish icecast connection.")
		initFn := func() error {
			// header
			meta := icecast.IcyMeta{Bitrate: int(bitrate), Channels: 2, SampleRate: sampleRate}

			header, err := icecast.GetIcecastPutHeader(icyAdd, meta, HTTP_VERSION)
			if err != nil {
				os.Exit(2)
			}
			if !network.WriteHeader(socketConn, header, 3, HTTP_VERSION, false) {
				return errors.New("could not send header")
			}
			return nil
		}
		// send buffer
		connection.StreamBuffer(socketConn, data, bitrate, 1024, reconnect, initFn, metaSendFn, nil)
	} else {
		log.Info("Establish connection.")
		connection.StreamBuffer(socketConn, data, bitrate, 1024, reconnect, nil, nil, nil)
	}

	wg.Done()
}

func GetMetaRequest(metaUrl, metaBody string) *http.Request {
	if metaUrl == "" {
		return nil
	}

	body := data.GetPayload(metaBody)
	if body == nil {
		return nil
	}

	// 	def to_local(ts):
	// 	return datetime.utcfromtimestamp(ts).strftime("%d.%m.%Y %H:%M:%S")
	// def to_iso(ts):
	// 	return datetime.utcfromtimestamp(ts).strftime("%Y-%m-%dT%H:%M:%SZ")

	now := time.Now().UTC()
	utcString := now.Format("2006-02-01T15:04:05Z")
	isoTimeString := now.Format("01.02.2006 15:04:05")
	bodyString := string(body)
	bodyString = strings.ReplaceAll(bodyString, "{{ start_utc }}", utcString)
	bodyString = strings.ReplaceAll(bodyString, "{{ start_iso }}", isoTimeString)

	// get request
	metaRequest, err := http.NewRequest(http.MethodPost, metaUrl, strings.NewReader(bodyString))
	if err != nil {
		log.Err(err, "create meta request")
	}
	return metaRequest
}

func Receive(validate string, metrics util.MetricsControl, verbose bool) {
	// if strings.ToLower(validate) == "audio" {
	// 	log.Newline()
	// 	log.Info("### Audio validation")
	// 	expectations := audio.Expectations{
	// 		IsCBR: true,
	// 		Encoding: audio.Encoding{
	// 			Bitrate:  256,
	// 			IsStereo: true,
	// 		},
	// 	}
	// 	expectations.Print()
	// 	log.Info("###")
	// 	log.Newline()
	// 	validator := audio.NewEncodingValidator(true, expectations, metricManager, verbose)
	// 	network.ReadTest(9999, true, validator, metricManager)
	// } else if strings.ToLower(validate) == "privatebit" {
	// 	log.Info("PrivateBit validation.")
	// 	validator := audio.NewPrivateBitValidator(true, audio.AudioTypeMp3, metricManager, verbose)
	// 	network.ReadTest(9999, true, validator, metricManager)
	// } else {
	// 	validator := network.DummyValidator{}
	// 	network.ReadTest(9999, true, validator, metricManager)
	// }
	// wg.Done()
}
