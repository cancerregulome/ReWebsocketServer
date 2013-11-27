//sockdrawer is a global singleton object for conectiong a client to a websocket and routing recived messages
//to the appropriate place.
//TODO: make this a non global singelton to support multiple conections per client?
var sockdrawer = {
	listenrawer = null
	listenaller = null
	routeingtable = [],
	ws = null
	
	//Connect is used to initialize the websocket conection by username. Only one such conection is supported at
	//a time.
	Conect = function(username) {
	  
	   if (ws != null) {
	     sockdrawer.ws.close();
	     sockdrawer.ws = null;
	   }
	 
	   
	   var url="ws"+location.origin.slice(4)+"/streamer/"+username;  
	  
	   sockdrawer.ws = new WebSocket(url);

	   sockdrawer.ws.onmessage = function (e) {
			try {
				if (sockdrawer.listenrawer !== null) {
					sockdrawer.listenrawer(e)
				}

				//TODO: check if valid json first instead of letting error fall through

				msg = JSON.parse(e)

				if (sockdrawer.listenaller !== null) {
					sockdrawer.listenaller(msg)
				}
				
				if ("conversationid" in msg) {
					sockdrawer.routeingtable[int(msg.conversationid)](msg)
				}

			} catch(err) {
				console.log(err)
			}
	   };

	},

	//Listen attaches the suplied handler function to a new conversation
	//and returns the conversation id for use in initiating queries and analysis
	Listen = function(handler) {
		sockdrawer.routeingtable.push(handler);
		return sockdrawer.routeingtable.length()-1;
	},

	//ListenAll attaches a handler that will recieve all messages as parsed json. It will replace
	//the previous handler.
	ListenAll = function(handler) {
		sockdrawer.listenaller=handler
	}

	//ListenRawer attaches a handler that will recieve all messages as raw strings It will replace
	//the previous handler.
	ListenRaw = function(handler) {
		sockdrawer.listenrawer=handler
	}




}