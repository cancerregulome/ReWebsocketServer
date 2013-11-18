package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"io"
	"os"
)

var host *string = flag.String("hostname", "localhost:23456", "Host:Port")
var username *string = flag.String("username", "user", "Username to conect as.")

//main method, parse input, and setup webserver
func main() {
	flag.Parse()

	prot := "ws"

	url := fmt.Sprintf("%v://%v/streamer/%v", prot, *host, *username)
	fmt.Println("Dialing Web Socket to ", url)

	var err error
	origin, err := os.Hostname()

	if err != nil {
		fmt.Println(err)
	}

	ws, err := websocket.Dial(url, "", fmt.Sprintf("http://%v", origin))
	if err != nil {
		fmt.Println(err)
	}

	_, err = io.Copy(os.Stdout, ws)
	if err != nil {
		fmt.Println(err)
	}

}
