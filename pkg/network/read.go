package network

import (
	"bufio"
	"flag"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nice-pink/goutil/pkg/log"
)

type Validator interface {
	Validate(data []byte, failEarly bool) error
}

// read stream

func ReadStream(url string, maxBytes uint64, outputFilepath string, reconnect bool, timeout time.Duration, dataValidator Validator, verbose bool) {
	// early exit
	if url == "" {
		log.Info()
		log.Error("Define url!")
		flag.Usage()
		os.Exit(2)
	}

	// log infos
	log.Info("Connect to url", url)
	if maxBytes > 0 {
		log.Info("Should stop after reading bytes:", maxBytes)
	} else {
		log.Info("Will read data until connection breaks.")
	}
	if outputFilepath != "" {
		log.Info("Dump data to file:", outputFilepath)
	}

	log.Info()
	iteration := 0
	for {
		log.Info("Start connection", iteration)
		log.Time()
		filepath := getFilePath(outputFilepath, iteration, true)
		ReadLineByLine(url, filepath, timeout, maxBytes, "", dataValidator)
		log.Time()
		log.Info()
		if !reconnect {
			break
		}
		iteration++
	}
}

func ReadLineByLine(url string, dumpToFile string, timeout time.Duration, maxBytes uint64, bearerToken string, dataValidator Validator) error {
	// request
	// build request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Err(err, "Request error.")
		return err
	}

	// auth
	if bearerToken != "" {
		var bearer = "Bearer " + bearerToken
		req.Header.Add("Authorization", bearer)
	}

	// request
	client := &http.Client{Timeout: timeout * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Err(err, "Client error.")
		return err
	}
	defer resp.Body.Close()

	// open file
	writeToFile := false
	var file *os.File = nil
	if dumpToFile != "" {
		file, err = os.Create(dumpToFile)
		writeToFile = true
		defer func() {
			if err := file.Close(); err != nil {
				log.Err(err, "Could not close file.")
			}
		}()
	}

	// read data
	var bytesRead uint64
	reader := bufio.NewReader(resp.Body)
	for {
		line, _ := reader.ReadBytes('\n')
		if writeToFile {
			file.Write(line)
		}

		// validate
		if dataValidator != nil {
			validationErr := dataValidator.Validate(line, true)
			if validationErr != nil {
				return validationErr
			}
		}

		bytesRead += uint64(len(line))

		if maxBytes > 0 && bytesRead > uint64(maxBytes) {
			log.Info()
			log.Info("Stop: Max bytes read", bytesRead)
			break
		}
	}

	return err
}

func getFilePath(baseFilePath string, iteration int, addTimestamp bool) string {
	if baseFilePath == "" {
		return ""
	}
	if addTimestamp {
		now := time.Now()
		baseFilePath += "_" + strconv.FormatInt(now.Unix(), 10)
	}
	return baseFilePath + "_" + strconv.Itoa(iteration)
}

// quick read test

func ReadTest(port int, printRate bool) {
	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		log.Err(err, "Listen error.")
		return
	}
	defer listener.Close()

	log.Info("Server is listening on port", port)

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