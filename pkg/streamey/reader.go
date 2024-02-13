package streamey

import (
	"net"
	"strconv"
	"time"

	"github.com/nice-pink/goutil/pkg/log"
)

func ReadTest(port int, printRate bool) {
	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		log.Err(err, "Listen error.")
		return
	}
	defer listener.Close()

	log.Info("Server is listening on port 8080")

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			log.Err(err, "Error:")
			continue
		}

		// Handle client connection in a goroutine
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	var readTotal int64 = 0
	readStart := time.Now().UnixNano()
	// Create a buffer to read data into
	buffer := make([]byte, 1024)

	var count int = 0
	for {
		// Read data from the client
		n, err := conn.Read(buffer)
		if err != nil {
			log.Err(err, "Read error.")
			return
		}
		readTotal += int64(n)

		// Process and use the data (here, we'll just print it)
		if count > 20 {
			rate := ((float64)(readTotal * 8)) / ((float64)(time.Now().UnixNano()-readStart) / 1000000000)
			log.Info("Current rate:", rate)
			count = 0
		}

		count++
	}
}
