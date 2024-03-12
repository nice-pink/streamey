package network

import (
	"net"
	"strings"
	"time"

	"github.com/nice-pink/goutil/pkg/log"
	"github.com/nice-pink/streamey/pkg/metricmanager"
)

func StreamBuffer(address string, sendBitRate float64, headerBuffer []byte, buffer []byte, reconnect bool, verbose bool) {
	log.Info("Stream data to", address, "with bitrate", sendBitRate, ". Reconnect", reconnect)
	for {
		// connection
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Error(err, "Can't dial.")
			return
		}
		defer conn.Close()

		headerWriteRetry := 5
		if len(headerBuffer) > 0 {
			writeHeader(conn, headerBuffer, headerWriteRetry, false)
		}

		// variables
		var bytesWrittenCycle int = 0
		var bytesWrittenTotal int64 = 0
		streamStart := time.Now().UnixNano()
		var byteIndex int64 = 0
		var byteSegmentSize int64 = 1024
		bufferLen := len(buffer)
		loopCount := 0

		// run loop
		var max int
		var dist int
		var count int = 1
		for {
			if byteIndex >= int64(bufferLen) {
				if verbose {
					log.Info("Start loop", loopCount)
				}
				byteIndex = 0
				count = 1
				loopCount++
				metricmanager.IncWriteLoopCounter()
			}
			/*
			 * calculate our instant rate over the entire transmit
			 * duration
			 */
			rate := ((float64)(bytesWrittenTotal * 8)) / ((float64)(time.Now().UnixNano()-streamStart) / 1000000000)

			// compare rate
			if rate < sendBitRate {
				max = min(bufferLen, count*int(byteSegmentSize))
				dist = max - int(byteIndex)
				// send data
				bytesWrittenCycle, err = conn.Write(buffer[byteIndex:max])
				if err != nil {
					log.Error(err, "Could not send data.")
					break
				}
				if bytesWrittenCycle <= 0 {
					log.Error("Is", bytesWrittenCycle)
				}
				if bytesWrittenCycle != dist {
					log.Error("Not all bytes sent. Should", dist, ", did", bytesWrittenCycle)
				}
				metricmanager.IncBytesWrittenCounter(bytesWrittenCycle)
				bytesWrittenTotal += int64(bytesWrittenCycle)
				byteIndex += int64(bytesWrittenCycle)

				count++
			}
		}

		// final log
		streamStop := time.Now().UnixNano()
		passed := streamStart - streamStop
		log.Info("Stopped sending. Bytes:", bytesWrittenTotal, ". Seconds:", passed)

		if !reconnect {
			break
		}
	}
}

func writeHeader(conn net.Conn, headerBuffer []byte, retry int, validate bool) bool {
	for counter := 0; counter < retry; counter++ {
		n, err := conn.Write(headerBuffer)
		if err != nil {
			log.Err(err, "Could not send data.")
			return false
		}

		if n < len(headerBuffer) {
			log.Error("Did not send entire header.")
			return false
		}
		log.Info("Header written", n)

		if !validate {
			return true
		}

		isValid := validateResponse(conn)
		if isValid {
			return true
		}
	}
	return false
}

func validateResponse(conn net.Conn) bool {
	// read and validate response
	var data []byte
	for {
		n, err := conn.Read(data)
		if err != nil {
			log.Err(err, "Read data from socket.")
			return true
		}

		if n > 0 {
			return isValidResponse(data)
		}

		log.Info("Read header response bytes", n)
		time.Sleep(time.Duration(2) * time.Second)
	}

}

func isValidResponse(data []byte) bool {
	dataString := string(data[:])
	if !strings.HasPrefix(dataString, "HTTP/1.1") {
		log.Error("Not a valid http response!")
		return false
	}

	split := strings.Split(dataString, "\r\n")
	if len(split) <= 0 {
		log.Error("No components in response.")
		return false
	}

	for _, key := range split {
		if key == "100 Continue" {
			return true
		}
	}

	log.Error("No 100 Continue!")
	return false
}
