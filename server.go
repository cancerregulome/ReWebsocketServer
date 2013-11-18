package main

import (
	"code.google.com/p/go.net/websocket"
	"crypto/rand"
	"flag"
	"fmt"
	"net/http"
	"path"
	"sync"
	"time"
)

//var port *int = flag.Int("port", 23456, "Port to listen.")
var host *string = flag.String("hostname", "localhost:23456", "Host:Port")
var contentDir *string = flag.String("contentdir", "./html/", "Path to html content dir.")

//Global registry to keep track of where to send incoming results
//TODO: maps aren't threadsafe so could break with lots of concurent inserts and should be RW mutexed
var WsRegistry = map[string]chan string{}
var RegistryRWMutex sync.RWMutex

//Task struct used for json serialization and submission to golem
type Task struct {
	Count int
	Args  []string
}

//generate a unique random string
func UniqueId() string {
	subId := make([]byte, 16)
	if _, err := rand.Read(subId); err != nil {
		fmt.Println(err)
	}
	return fmt.Sprintf("%x", subId)
}

//Handeler for websocket connectionf from client pages. Expects a []int list of feature id's as its first message
//and will submit these tasks and then wait to stream results back
func SocketStreamer(ws *websocket.Conn) {
	fmt.Printf("jsonServer %#v\n", ws.Config())
	//for {
	url := *ws.Config().Location
	id := path.Base(url.Path)

	RestultChan := make(chan string, 0)
	RegistryRWMutex.Lock()
	WsRegistry[id] = RestultChan
	RegistryRWMutex.Unlock()

	for {
		msg := <-RestultChan
		err := websocket.Message.Send(ws, msg)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("send:%#v\n", msg)
	}

	//TODO: remove ws from registry and cancel outstandign jobs when connetion dies
}

//excepts posts from tasks and dispatches them to appropriate Socket Streamer
func ResultHandeler(w http.ResponseWriter, req *http.Request) {
	//TODO: check to make sure it is a post with results in it
	//and sanatize it so people can't push arbitray stuff to the client
	fmt.Println("result request ", req.URL.Path)
	id := path.Base(req.URL.Path)

	RegistryRWMutex.RLock()
	val, ok := WsRegistry[id]
	RegistryRWMutex.RUnlock()

	if ok {
		select {
		case val <- req.FormValue("results"):

		case <-time.After(5 * time.Second):
			fmt.Println("timed out streaming to task ", id)
		}
	}
}

//main method, parse input, and setup webserver
func main() {
	flag.Parse()
	fmt.Println("Serving ", *host)
	fmt.Printf("Conect websockets to %v/streamer/id and post messages to %v/results/id\n\n", *host, *host)
	http.Handle("/streamer/", websocket.Handler(SocketStreamer))
	http.HandleFunc("/results/", ResultHandeler)
	http.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir(*contentDir))))
	http.Handle("/", http.RedirectHandler("/html/index.html", http.StatusTemporaryRedirect))
	//fmt.Printf("http://localhost:%d/\n", *port)
	err := http.ListenAndServe(*host, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
