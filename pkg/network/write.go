package network

import (
	"net"
	"time"

	"github.com/nice-pink/goutil/pkg/log"
)

func StreamBuffer(address string, sendBitRate float64, buffer []byte, reconnect bool) {
	for {

		// connection
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Error(err, "Can't dial.")
			return
		}
		defer conn.Close()

		// variables
		var bytesWrittenCycle int = 0
		var bytesWrittenTotal int64 = 0
		streamStart := time.Now().UnixNano()
		var byteIndex int64 = 0
		var byteSegmentSize int64 = 1024
		bufferLen := len(buffer)
		loopCount := 0

		// run loop
		var count int = 1
		for {
			if byteIndex >= int64(bufferLen)-1 {
				byteIndex = 0
				count = 1
				loopCount++
				log.Info("Start loop", loopCount)
			}
			/*
			 * calculate our instant rate over the entire transmit
			 * duration
			 */
			rate := ((float64)(bytesWrittenTotal * 8)) / ((float64)(time.Now().UnixNano()-streamStart) / 1000000000)

			// compare rate
			if rate < sendBitRate {
				// send data
				bytesWrittenCycle, err = conn.Write(buffer[byteIndex:min(bufferLen-1, count*int(byteSegmentSize))])
				if err != nil {
					log.Error(err, "Could not send data.")
					break
				}
				if bytesWrittenCycle <= 0 {
					log.Error("Is", bytesWrittenCycle)
				}
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
