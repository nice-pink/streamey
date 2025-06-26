package streamer

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nice-pink/audio-tool/pkg/network"
	"github.com/nice-pink/audio-tool/pkg/stream"
	"github.com/nice-pink/audio-tool/pkg/util"
	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/configmanager"
	"github.com/nice-pink/streamey/pkg/metadata"
)

const (
	HTTP_VERSION string = "1.0"
	CHUNK_SIZE   int    = 1024
)

func Stream(config configmanager.StreamConfig, metrics util.MetricsControl, wg *sync.WaitGroup, verbose bool) {
	// read file
	// TODO - get correct item from playlist
	filepath := config.Playlist.Items[0].Filepath
	data := getData(filepath)
	if len(data) == 0 {
		log.Error("no data in file", filepath)
		os.Exit(2)
	}

	// get url
	streamFormat := configmanager.GetStreamFormat(config.Audio.Format)
	url, connTarget := getUrlAndTarget(config.Audio.TargetUrl, streamFormat)
	log.Info("Stream data with bitrate", config.Audio.Bitrate, "to", url)

	// stream
	log.Info("Conn to url", url)
	port := 80
	if strings.HasPrefix(url, "https://") {
		port = 443
	}
	connection := network.NewConnection(url, "", port, 0, time.Duration(30), network.HttpConnection, metrics)
	connection.VerboseLogs = verbose
	defer connection.Close()

	_, err := connection.GetSocketConn()
	if err != nil {
		log.Err(err, "cannot establish socket connection to", url)
		os.Exit(2)
	}

	// send metadata
	httpClient := http.Client{}
	metaSendFn := func(loopCount int) error {
		metaRequest := metadata.GetMetadataRequest(config.Metadata.TargetUrl, config.Metadata.Template, config.Playlist.ContentType, config.Metadata.Headers, config.Playlist.Items, loopCount, true)
		if metaRequest == nil {
			return nil
		}
		resp, err := httpClient.Do(metaRequest)
		if err != nil {
			log.Err(err, "send metadata error")
		}
		if resp.StatusCode >= 300 {
			log.Error("metadata request: status code >= 300:", resp.StatusCode)
		}
		return err
	}

	// init function
	if streamFormat == configmanager.StreamFormatIcecast || streamFormat == configmanager.StreamFormatShoutcast {
		log.Info("Establish icecast connection.")
		initFn := func() error {
			// header
			var header []byte
			var err error
			switch streamFormat {
			case configmanager.StreamFormatIcecast:
				meta := stream.IcyMeta{Bitrate: int(config.Audio.Bitrate), Channels: 2, SampleRate: config.Audio.SampleRate, Url: config.Audio.TargetUrl}
				header, err = stream.GetIcecastPutHeader(connTarget, meta, HTTP_VERSION, false)
			case configmanager.StreamFormatShoutcast:
				header, err = stream.GetShoutcastSourceHeader(connTarget, HTTP_VERSION, false)
			}

			if err != nil {
				os.Exit(2)
			}

			// log get socket conn
			conn, err := connection.GetSocketConn()
			if err != nil {
				log.Err(err, "no socket connection in initFn")
				return err
			}
			if !network.WriteHeader(conn, header, 3, HTTP_VERSION, true, false) {
				return errors.New("could not send header")
			}
			return nil
		}
		// send buffer
		connection.StreamBuffer(data, float64(config.Audio.Bitrate), CHUNK_SIZE, true, initFn, metaSendFn, nil)
	} else {
		log.Info("Establish connection.")
		connection.StreamBuffer(data, float64(config.Audio.Bitrate), CHUNK_SIZE, true, nil, nil, nil)
	}

	wg.Done()
}

// helper

func getUrlAndTarget(targetUrl string, streamFormat configmanager.StreamFormat) (string, stream.ConnTarget) {
	url := targetUrl
	var connTarget stream.ConnTarget
	var err error
	if streamFormat == configmanager.StreamFormatIcecast || streamFormat == configmanager.StreamFormatShoutcast {
		// overwrite url with icy address
		connTarget, err = stream.GetConnTarget(url)
		if err != nil {
			log.Error("Icy address")
			os.Exit(2)
		}
		connTarget.UserAgent = "streamey/1.0"
		url = connTarget.Domain
	}
	return url, connTarget
}

func getData(filepath string) []byte {
	log.Info("Get data from", filepath)
	filepathFinal := filepath
	if strings.HasPrefix(filepath, "http") {
		filepathFinal = "file.mp3"
		util.DownloadFile(filepath, filepathFinal)
	}

	// is local file
	file, err := os.Open(filepathFinal)
	if err != nil {
		log.Err(err, "Cannot open file.", filepathFinal)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		log.Err(err, "Cannot read file.", filepathFinal)
	}
	return data
}

// test

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
