ReWebSocketServer
=================

This is a simple prototype and demo of a server for streaming messages to Reglome Explorer users via websockets.

Each user connects a websocket to /streamer/uniqueid and server side services can post messages to
/results/uniqueid

POST request should have the message in the "results" form value. 

Multipart request are not supported but can be.

Websocket server based on code from the google io demo.


Quick Start
------------


```
#download and build
git clone https://github.com/cancerregulome/ReWebsocketServer.git
cd ReWebsocketServer
go get
go build

#start server
./ReWebSocketServer -hostname="localhost:23456" -contentdir="./html/"

```

Using and Testing
------------------

The number of concurent clients is limited by the number of open files (ulimit -n). 


```
#open localhost:23456 in browser and conect a user named "user" (or whatever)

#post a test message to  user
cd testing
python poster.py testmsg.txt http://localhost:23456/results/user

#build test client in testing/testclient
cd testclient
go build

#connect 100 test clients named client0-client99
bash conectn.sh 200 localhost:23456

```


