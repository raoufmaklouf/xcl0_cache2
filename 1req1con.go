package main

//var tlsConn *tls.Conn

func normalRequest(method string, path string, host string, port int) string {
	//host := "0a1000ce03b1f1e883ea4b4c004800af.web-security-academy.net"
	//port := 443
	var responses1 string = ""

	// Create a TCP connection
	tcpConn, err := createTCPConnection(host, port)
	if err == nil {
		defer tcpConn.Close()
		tlsConn, err = createTLSConnection(tcpConn)
		if err == nil {
			defer tlsConn.Close()
			///robots.txt
			err = sendRequest("%s %s HTTP/1.1\r\nHost: %s\r\nUser-Agent: Mozilla/5.0\r\n\r\n", method, path, host)
			if err == nil {
				responsePrefix := "HTTP/1.1"
				responseCount := 2
				combinedResponse, err := readFullResponse(responsePrefix, responseCount)
				if err == nil {
					responses1 = combinedResponse

				}

			}

		}

	}
	//fmt.Println(responses1)

	return responses1
}
