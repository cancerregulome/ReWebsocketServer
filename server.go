package main

import (
	"code.google.com/p/go.net/websocket"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

//var port *int = flag.Int("port", 23456, "Port to listen.")
var host *string = flag.String("hostname", "localhost:23456", "Host:Port")
var contentDir *string = flag.String("contentdir", "./html/", "Path to html content dir.")
var cookie_secret *string = flag.String("cookie_secret", "", "Secret used to validate secure cookies.")

/*get_secure_cookie is a port of tornado webs cookie verificaiton algorythem for authenticating client conections.
From tornado:

	def decode_signed_value(secret, name, value, max_age_days=31):
	    if not value:
	        return None
	    parts = utf8(value).split(b"|")
	    if len(parts) != 3:
	        return None
	    signature = _create_signature(secret, name, parts[0], parts[1])
	    if not _time_independent_equals(parts[2], signature):
	        gen_log.warning("Invalid cookie signature %r", value)
	        return None
	    timestamp = int(parts[1])
	    if timestamp < time.time() - max_age_days * 86400:
	        gen_log.warning("Expired cookie %r", value)
	        return None
	    if timestamp > time.time() + 31 * 86400:
	        # _cookie_signature does not hash a delimiter between the
	        # parts of the cookie, so an attacker could transfer trailing
	        # digits from the payload to the timestamp without altering the
	        # signature.  For backwards compatibility, sanity-check timestamp
	        # here instead of modifying _cookie_signature.
	        gen_log.warning("Cookie timestamp in future; possible tampering %r", value)
	        return None
	    if parts[1].startswith(b"0"):
	        gen_log.warning("Tampered cookie %r", value)
	        return None
	    try:
	        return base64.b64decode(parts[0])
	    except Exception:
	        return None


	def _create_signature(secret, *parts):
	    hash = hmac.new(utf8(secret), digestmod=hashlib.sha1)
	    for part in parts:
	        hash.update(utf8(part))
	    return utf8(hash.hexdigest())

*/
func get_secure_cookie(ws *websocket.Conn, name string, secret string) string {

	cookie, err := ws.Request().Cookie(name)
	if err != nil {
		return ""
	}
	value := cookie.String()
	if len(value) == 0 {
		return ""
	}
	parts := strings.Split(value, "|")
	if len(parts) != 3 {
		return ""
	}
	hash := hmac.New(sha1.New, []byte(secret))

	_, _ = hash.Write([]byte(parts[0]))

	if !hmac.Equal(hash.Sum([]byte(parts[1])), []byte(parts[2])) {
		return ""
	}

	if cookie.Expires.Before(time.Now().Add(time.Hour * -24 * 31)) {
		return ""
	}
	if cookie.Expires.After(time.Now().Add(time.Hour * 24 * 31)) {
		return ""
	}

	if []byte(parts[1])[0] == byte(0) {
		return ""
	}

	data, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return ""
	}
	return string(data)

}

//Global registry to keep track of where to send incoming results
var WsRegistry = map[string]chan string{}
var RegistryRWMutex sync.RWMutex

//Handeler for websocket connectionf from client pages.
func SocketStreamer(ws *websocket.Conn) {

	//for {
	url := *ws.Config().Location
	id := path.Base(url.Path)

	fmt.Printf("Socket conectiong to %#v\n", url)

	if len(*cookie_secret) > 0 {
		user := get_secure_cookie(ws, "woami", *cookie_secret)
		if len(user) == 0 {
			ws.Close()
			return
		}
	}

	RestultChan := make(chan string, 0)
	RegistryRWMutex.Lock()
	WsRegistry[id] = RestultChan
	RegistryRWMutex.Unlock()

	for {

		select {
		case msg := <-RestultChan:
			err := websocket.Message.Send(ws, msg)
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Printf("send:%#v\n", msg)

		case <-time.After(5 * time.Minute):
			fmt.Println("timed out streaming to user ", id)
			break
		}

	}

	RegistryRWMutex.Lock()
	delete(WsRegistry, id)
	RegistryRWMutex.Unlock()
	ws.Close()

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
			fmt.Println("timed out streaming to user ", id)
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
