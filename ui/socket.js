let port = location.port
if (port == "") {
  if (location.protocol == "https:") {
    port = 443
  }else {
    port = 80
  }
}
const baseUrl = window.location.hostname+':'+port
let webSocketType = "ws"
if (location.protocol == "https:") {
  webSocketType = "wss"
}
socket = null

function startWebsocket() {
    socket = new WebSocket(webSocketType+'://'+baseUrl+"/ws");

    // Connection opened
    socket.addEventListener('open', () => {
      console.log("connection established")
      //let body = JSON.stringify({slide: this.state.page, code: codeText})
      //console.log(body)
      //socket.send(body);
    });

    // Listen for messages
    socket.addEventListener('message', (event) => {
        console.log('Message from server: ', event.data);
        const data = JSON.parse(event.data)
        if (data.Data != "" && data.Data != null) {
          if (data.Pool != '') {
            updateGraph(data.Pool, data.Data)
          }
          return
        }
        if (data.Reload) {
          location.reload()
        } else {
          if (data.Slide > -999){
            console.log(data)
            if (data.ID == "00000000000000000000000000 "){
              return
            }
            if (myID == ""){
              myID = data.ID
              console.log("myID",myID)
            }
            if (data.Slide != page){
              setPage(data.Slide)
            }
          }
        }
    });

    socket.onclose = () => {
      console.log('Socket is closed');
      socket.close();
      socket = null
      myID = ""
      setTimeout(startWebsocket, 5000)
    };
  }

  function updateData(data) {
    console.log("updateData send", data)

    if (socket === null) {
      console.log("socket is null, skipping send")
      return
    }
    if (socket.readyState === WebSocket.CLOSED || socket.readyState === WebSocket.CLOSING) {
      console.log("WebSocket is closed or closing, skipping send")
      return
    }
    socket.send(JSON.stringify(data))
};
