package metadata

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nice-pink/goutil/pkg/data"
	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/configmanager"
)

type MetaTypeId int

const (
	MetaTypeIdSong       MetaTypeId = 1
	MetaTypeIdSpot       MetaTypeId = 2 // ???
	MetaTypeIdLink       MetaTypeId = 3 // ad
	MetaTypeIdVoiceTrack MetaTypeId = 4 // ???
	MetaTypeIdAd         MetaTypeId = 5
)

func GetMetadataRequest(metaUrl, metaBody, contentType string, headers map[string]string, items []configmanager.PlaylistItem, loopCount int, isInit bool) *http.Request {
	if metaUrl == "" || len(items) == 0 {
		return nil
	}

	// get meta body from item
	lenItems := len(items)
	index := loopCount % lenItems
	body := GetMetadataBody(metaBody, items[index], isInit)

	// get request
	metaRequest, err := http.NewRequest(http.MethodPost, metaUrl, bytes.NewReader(body))
	if err != nil {
		log.Err(err, "create meta request")
	}

	// headers
	metaRequest.Header.Add("content-type", contentType)
	for k, v := range headers {
		metaRequest.Header.Add(k, v)
	}

	return metaRequest
}

func GetMetadataBody(metaBody string, item configmanager.PlaylistItem, lastStarted bool) []byte {
	body := data.GetPayload(metaBody)
	if body == nil {
		return nil
	}

	bodyString := string(body)

	// type
	bodyString = strings.ReplaceAll(bodyString, "{{ type_name }}", item.Type)
	typeId := GetMetaTypeId(item.Type)
	bodyString = strings.ReplaceAll(bodyString, "{{ type_id }}", typeId)

	// ids
	bodyString = strings.ReplaceAll(bodyString, "{{ uuid }}", uuid.NewString())
	bodyString = strings.ReplaceAll(bodyString, "{{ id }}", "1")

	// event
	bodyString = strings.ReplaceAll(bodyString, "{{ last_started }}", strconv.FormatBool(lastStarted))

	// track
	bodyString = strings.ReplaceAll(bodyString, "{{ artist }}", item.Artist)
	bodyString = strings.ReplaceAll(bodyString, "{{ title }}", item.Title)
	bodyString = strings.ReplaceAll(bodyString, "{{ album }}", item.Album)
	bodyString = strings.ReplaceAll(bodyString, "{{ filepath }}", item.Filepath)
	bodyString = strings.ReplaceAll(bodyString, "{{ duration }}", strconv.FormatFloat(item.Duration, 'f', 3, 64))

	// times
	now := time.Now().UTC()
	utcFormat := "2006-01-02T15:04:05Z"
	isoFormat := "02.01.2006 15:04:05"
	// start
	bodyString = strings.ReplaceAll(bodyString, "{{ start_utc }}", now.Format(utcFormat))
	bodyString = strings.ReplaceAll(bodyString, "{{ start_iso }}", now.Format(isoFormat))
	// stop
	stop := now.Add(time.Duration(item.Duration) * time.Second)
	bodyString = strings.ReplaceAll(bodyString, "{{ stop_utc }}", stop.Format(utcFormat))
	bodyString = strings.ReplaceAll(bodyString, "{{ stop_iso }}", stop.Format(isoFormat))

	return []byte(bodyString)
}

func GetMetaTypeId(typeName string) string {
	id := MetaTypeIdSong
	switch strings.ToLower(typeName) {
	case "spot":
		id = MetaTypeIdSpot
	case "link":
		id = MetaTypeIdLink
	case "voicetrack":
		id = MetaTypeIdVoiceTrack
	case "ad":
		id = MetaTypeIdAd
		// default:
		// 	id = MetaTypeIdSong
	}
	return strconv.Itoa(int(id))
}
