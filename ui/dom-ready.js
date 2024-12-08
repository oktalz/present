document.addEventListener("DOMContentLoaded", function() {
  setPage(page);
  document.querySelectorAll('pre code.code-edit').forEach(function(codeElement) {
    codeElement.contentEditable = "true";
    codeElement.spellcheck = false;
  });
  document.querySelectorAll('code').forEach(function(codeElement) {
    if (codeElement.classList.length == 0) {
        codeElement.classList.add("hljs")
        //codeElement.style.color = "#00ADD8" go colors
        codeElement.style.color = "#800"
        codeElement.style.backgroundColor = "#f8f8f8";
    }
  });
  startSSESession();
  notesPause = false
  notesCounter = 0

  // Get all elements with id 'd2-svg'
  var svgElements = document.querySelectorAll('#d2-svg');
  // Loop through each of them
  svgElements.forEach(function(svg) {
    correctD2Graph(svg);
  });

  if (window.location.search.indexOf('notes') != -1) {
    document.getElementById("presenter-time-top").classList.remove("closed");
    document.getElementById("presenter-current-time").classList.remove("closed");
    elements = document.getElementsByClassName("presenter-comment");
    for (let i = 0; i < elements.length; i++) {
      elements[i].classList.remove("closed");
    }
    setInterval(function() {
      let minutes = Math.floor(notesCounter / 60);
      let seconds = notesCounter % 60;
      if (!notesPause) {
        notesCounter++;
        value = (minutes < 10 ? "0" : "") + minutes + ":" + (seconds < 10 ? "0" : "") + seconds
        document.getElementById("presenter-time").innerHTML = value;
      }

      let now = new Date();
      let hours = now.getHours();
      minutes = now.getMinutes();
      seconds = now.getSeconds();
      value = (hours < 10 ? "0" : "") + hours + ":" + (minutes < 10 ? "0" : "") + minutes
      document.getElementById("presenter-current-time").innerHTML = value;
    }, 1000);
  }
});
