function startSSESession() {
    const eventSource = new EventSource('/events');
    // Listen for messages from the server
    eventSource.onmessage = function(event) {
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
    };
    // Handle errors
    eventSource.onerror = function(event) {
      console.error('Error occurred:', event);
     };
}

  function updateData(data) {
    fetch('/events', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    })
    .catch(error => {
      console.error('Error occurred when sending data:', error);
    })
    console.log("updateData send", data)
};
