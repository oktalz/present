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

    let data = {slide: pageNum, code: codeText}
    if (block != -1) {
      data.block= block
    }

    ////////////////////////////////
    const xhr = new XMLHttpRequest();
    xhr.open('POST', '/cast');
    xhr.responseType = 'text';

    xhr.onreadystatechange = function() {
    if (xhr.readyState === 4) {
        if (xhr.status === 200) {
          //console.log(xhr.responseText);
        } else {
          console.error('Request failed:', xhr.statusText);
        }
    }
    };

    xhr.onprogress = function(event) {
        if (event.target.readyState === XMLHttpRequest.DONE) {
            return;
        }
        const data = event.target.responseText;
        terminalElement.innerHTML = data
        terminalElement.scrollTop = terminalElement.scrollHeight
    };
    xhr.onloadend = function() {
        if (xhr.status === 200) {
          //console.log('Request completed successfully:', xhr.responseText);
        } else {
          console.error('Request failed:', xhr.statusText);
        }
        setSpinner(false)
      };

    xhr.setRequestHeader('Content-Type', 'application/json');  // Adjust based on your data format
    xhr.send(JSON.stringify(data));
}
