function castTerminal(block){
    pageNum = page

    const slideElement = document.getElementById('slide-'+pageNum);
    const terminalElement = document.getElementById('terminal-'+pageNum);
    const terminalElementX = document.getElementById('terminalx-'+pageNum);
    if (terminalElement == null) {
        return
    }
    setSpinner(true)
    terminalElement.innerHTML = ''
    terminalElement.classList.remove('closed');
    terminalElementX.classList.remove('closed');
    codeText = ''
    if (slideElement) {
        let codeElements = slideElement.querySelectorAll('pre code.code-cast');
        codeText = Array.from(codeElements).map(codeElement => codeElement.innerText);
    }

    const socket = new WebSocket(webSocketType+'://'+baseUrl+"/cast");

    // Connection opened
    socket.addEventListener('open', () => {
        let data = {slide: pageNum, code: codeText}
        if (block != -1) {
          data.block= block
        }
        let body = JSON.stringify(data)
        console.log(body)
        socket.send(body);
        //this.CheckTabState(codeText)
    });

    // Listen for messages
    socket.addEventListener('message', (event) => {
        console.log('Message from server: ', event.data);
        terminalElement.innerHTML += event.data + '<br>'
        terminalElement.scrollTop = terminalElement.scrollHeight;
    });

    socket.onclose = () => {
        console.log('Socket is closed');
        socket.close();
        setSpinner(false)
    };

}
