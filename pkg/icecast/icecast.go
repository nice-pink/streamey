package icecast

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/nice-pink/goutil/pkg/log"
)

type IcyMeta struct {
	Bitrate    int
	SampleRate int
	Channels   int
	Quality    int
	AudioInfo  string
}

type IcyAdd struct {
	Scheme     string
	Domain     string
	MountPoint string
	Port       string
	BasicAuth  string
}

func GetIcyAddress(fullUrl string) (IcyAdd, error) {
	full, err := url.Parse(fullUrl)
	if err != nil {
		return IcyAdd{}, err
	}

	domain := full.Hostname()
	if domain == "" {
		return IcyAdd{}, errors.New("invalid url")
	}

	scheme := full.Scheme
	mountPoint := full.Path
	port := full.Port()
	password, hasPassword := full.User.Password()
	basicAuth := ""
	if hasPassword {
		basicAuth = getBasicAuth(full.User.Username(), password)
	}
	return IcyAdd{Scheme: scheme, Domain: domain, MountPoint: mountPoint, Port: port, BasicAuth: basicAuth}, nil
}

func GetIcecastPutHeader(icyAdd IcyAdd, meta IcyMeta) ([]byte, error) {
	header := "PUT " + icyAdd.MountPoint + " HTTP/1.1\nHost: " + icyAdd.Domain + ":" + icyAdd.Port + "\n"
	if icyAdd.BasicAuth != "" {
		header += "Authorization: Basic " + icyAdd.BasicAuth + "\n"
	}
	header += addIcyMeta(meta)
	return convertToByteHeader(header, false), nil
}

func getBasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func addIcyMeta(meta IcyMeta) string {
	audioInfo := "samplerate=" + strconv.Itoa(meta.SampleRate) + ";quality=" + strconv.Itoa(meta.Quality) + ";channels=" + strconv.Itoa(meta.Channels)
	return `Content-Type: audio/mpeg
Accept: */*
User-Agent: streamey
Server: Icecast 2.4.0-kh15
icy-br:` + strconv.Itoa(meta.Bitrate) + `
icy-genre:Test
icy-name:SineSweep
icy-notice1:This is a radiosphere test stream.
icy-pub:0
icy-url:http://channel.radiosphere.io
Icy-MetaData:0
icy-audio-info:` + audioInfo + `
ice-audio-info:` + audioInfo + `
Expect: 100-continue
`
}

func convertToByteHeader(header string, print bool) []byte {
	header = strings.Replace(header, "\n", "\r\n", -1)
	header += "\r\n"
	if print {
		log.Info("Header:\n" + header)
		log.Info("Header size:", len(header))
	}
	return []byte(header)
}
