var currentPage = /^#?\d+$/.test(window.location.hash) ? parseInt(window.location.hash.slice(1), 10) : 0;
var previousPage = -100;
const pagesWithAutoPlayAudio = new Map();
const pagesWithAutoPlayAudioPlayed = new Map();

if (window.self !== window.top) {
  topPage = /^#?\d+$/.test(window.top.location.hash) ? parseInt(window.top.location.hash.slice(1), 10) : 0;
  currentPage = topPage
}
setPage(currentPage);
// window.addEventListener('hashchange', () => {
//   console.log('Hash changed (print):', window.location.hash);
// }, false);
var spinner = false
var myID = ""
var showMenu = false
var showOptions = false

function setSpinner(value){
    spinner = value
    if (value) {
        document.getElementById("run-"+currentPage+"-refresh").classList.remove("closed")
        document.getElementById("run-"+currentPage+"").classList.add("closed")
    } else {
        document.getElementById("run-"+currentPage+"-refresh").classList.add("closed")
        document.getElementById("run-"+currentPage+"").classList.remove("closed")
    }
}

function getSlideElements() {
  return document.querySelectorAll('[id^="slide-"]');
}

function setPageWithUpdate(newPage) {
  setPage(newPage);
  updateData({
    Author: myID,
    Slide: newPage
  })
}

function triggerPool(key, value) {
  console.log(key, value)
  updateData({
    Author: myID,
    Pool: key,
    Value: value
  })
}

function setPage(newPage) {
  if (newPage < -2){
    return
  }
  // remove class menu-selected from elements that has class menu-selected
  document.querySelectorAll('.menu-selected').forEach(el => {
    el.classList.remove('menu-selected');
  });
  if (previousPage != newPage && newPage > -1) {
    pauseAudio(previousPage)
    playAudioIfPaused(newPage, pagesWithAutoPlayAudio.get("audio-page-"+newPage))
  }

  previousPage = currentPage
  currentPage = newPage
  if (currentPage < 0) {
    currentPage = 0;
  }
  if (currentPage > maxPage) {
    currentPage = maxPage
  }
  //set class menu-selected to element that has id menu+page
  menu = document.getElementById('menu-'+currentPage)
  if (menu != null) {
    menu.classList.add('menu-selected')
  }
  window.location.hash = currentPage.toString();
  updateSlideVisibility();
  menu = document.getElementById('menu')
  if (menu != null) {
    menu.classList.add('menu-hidden')
    showMenu = false
  }
  if (typeof window.onPageChange === 'function') {
    window.onPageChange();
  } else {
    let intervalOnPageChange = setInterval(function() {
      if (typeof window.onPageChange === 'function') {
        window.onPageChange();
        clearInterval(intervalOnPageChange);
      }
    }, 10);
  }
}

function updateSlideVisibility() {
  getSlideElements().forEach(function(slide) {
    if (parseInt(slide.id.slice(6)) == currentPage) {
      slide.classList.remove('hidden');
      slide.classList.add('visible');
    } else {
      slide.classList.add('hidden');
      slide.classList.remove('visible');
    }
  });
}

function closeTerminal(){
  const terminalElement = document.getElementById('terminal-'+currentPage);
  const terminalElementX = document.getElementById('terminalx-'+currentPage);
  terminalElement.innerHTML = ''
  terminalElement.classList.add('closed');
  terminalElementX.classList.add('closed');
}

document.addEventListener('keydown', function(e) {
  var keyCode = e.key;
  activeElement = document.activeElement;
  if (activeElement != null) {
    if (keyCode == 'Escape' ) {
      activeElement.blur();
    }
    if (activeElement.classList.contains('hljs') && (activeElement instanceof HTMLElement && activeElement.isContentEditable)) {
        return;
    }
  }
  if (nextPageKeys.includes(keyCode)) {
    oldPage = currentPage;
    newPage = currentPage + 1;
    target = getPageUp(oldPage,newPage);
    setPage(target);
    updateData({
      Author: myID,
      Slide: currentPage
    })
  }
  if (previousPageKeys.includes(keyCode)) {
    oldPage = currentPage;
    newPage = currentPage - 1;
    target = getPageDown(oldPage,newPage);
    setPage(target);
    updateData({
      Author: myID,
      Slide: currentPage
    })
  }
  if (terminalCast.includes(keyCode)) {
    castTerminal(-1)
  }
  if (terminalClose.includes(keyCode)) {
    closeTerminal()
  }
  if (menuKey.includes(keyCode)) {
    showMenu = !showMenu
    if (showMenu) {
      document.getElementById('menu').classList.remove('menu-hidden');
      let targetElement = null
      let index = currentPage
      while (targetElement == null && index < maxPage) {
        targetElement = document.getElementById(`menu-`+index);
        if (targetElement) {
          targetElement.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
        index = index + 1
      }
    } else {
      document.getElementById('menu').classList.add('menu-hidden');
    }
  }
  if (optionsKey.includes(keyCode)) {
    showOptions = !showOptions
    if (showOptions) {
      document.getElementById('options').classList.remove('options-hidden');
    } else {
      document.getElementById('options').classList.add('options-hidden');
    }
  }
});

window.addEventListener('wheel', function(event) {
  activeElement = document.activeElement;
  if (activeElement != null) {
    if (activeElement.classList.contains('hljs') && (activeElement instanceof HTMLElement && activeElement.isContentEditable)) {
        return;
    }
  }
  const eventObj = event || window.event; // cross browser
  var target = eventObj.target
  var rect = target.getBoundingClientRect()
  x = eventObj.clientX - rect.left,
  y = eventObj.clientY - rect.top;
  var elementUnderMouse = document.elementFromPoint(x, y);
  if (elementUnderMouse != null) {
    if (elementUnderMouse.classList.contains('hljs') && (elementUnderMouse instanceof HTMLElement && elementUnderMouse.isContentEditable)) {
        return;
    }
  }
  const hoveredElements = document.querySelectorAll(':hover');
  menuElements = Array.from(hoveredElements).filter((el) => el.classList.contains('menu'));
  if (menuElements.length > 0) {
    return;
  }
  menuElements = Array.from(hoveredElements).filter((el) => el.classList.contains('box-overflow'));
  if (menuElements.length > 0) {
    return;
  }
  menuElements = Array.from(hoveredElements).filter((el) => el.classList.contains('terminal'));
  if (menuElements.length > 0) {
    return;
  }

  const scrollDirection = event.deltaY > 0 ? 'downward' : 'upward';
  //console.log(`Mouse scroll ${scrollDirection}: ${Math.abs(event.deltaY)} pixels`);
  if (scrollDirection == 'downward') {
      oldPage = currentPage;
      newPage = currentPage + 1;
      target = getPageUp(oldPage,newPage);
      setPage(target);
      updateData({
        Author: myID,
        Slide: currentPage
      })
  }
  if (scrollDirection == 'upward') {
      oldPage = currentPage;
      newPage = currentPage - 1;
      target = getPageDown(oldPage,newPage);
      setPage(target);
      updateData({
        Author: myID,
        Slide: currentPage
      })
  }
}, true);

window.addEventListener("popstate", function(e) {
  // detect that back is clicked and go to the appropriate page
  // add a func that activates after 50ms when this is activated (its to early to do it immediately)
  this.setTimeout(() => {
    //remove the hash from the URL
    hash = window.location.hash;
    if (hash != "") {
      hash = hash.slice(1);
      if (/^\d+$/.test(hash)) {
        newPage = parseInt(hash, 10);
        if (currentPage != newPage) {
          setPage(newPage);
        }
      }
    }
  }, 50);
  //console.log('Popstate event triggered:', e.state);
}, false);

touchX = 0;
touchY = 0;

document.addEventListener("pointerdown", (e) => {
  touchStartX = e.clientX;
  touchStartY = e.clientY;
});

document.addEventListener("pointermove", (e) => {
  touchEndX = e.clientX;
  touchEndY = e.clientY;
});

document.addEventListener('touchstart', function (event) {
  touchX = event.changedTouches[0].screenX;
  touchY = event.changedTouches[0].screenY;
}, false);

document.addEventListener('touchend', function (event) {
  endX = event.changedTouches[0].screenX;
  endY = event.changedTouches[0].screenY;
  if (Math.abs(endX - touchX) <= Math.abs(endY - touchY)) {
    return;
  }

  if (endX > touchX) {
    oldPage = currentPage;
    newPage = currentPage - 1;
    target = getPageDown(oldPage,newPage);
    setPage(target);
    updateData({
      Author: myID,
      Slide: currentPage
    })
  }

  if (endX < touchX) {
    oldPage = currentPage;
    newPage = currentPage + 1;
    target = getPageUp(oldPage,newPage);
    setPage(target);
    updateData({
      Author: myID,
      Slide: currentPage
    })
  }
}, false);


document.addEventListener('click', function(event) {
  var targetId = event.target.id;
  if (targetId && targetId.startsWith('run-icon-')) {
    var pageNumber = targetId.substring(9); // strip 'run-' prefix
    console.log(`Clicked run button for page ${pageNumber}`);
    castTerminal(-1)
  }
});

function tabChangeGlobal(tabID){
  console.log(tabID)
  let pageDIV = document.getElementById("slide-"+currentPage);
  let tablinks = Array.from(pageDIV.querySelectorAll(".tablinks"));
  console.log(tablinks)
  for (let i = 0; i < tablinks.length; i++) {
    if (tablinks[i] && tablinks[i].getAttribute('id') === "tab-"+tabID) {
      tablinks[i].classList.add("active");
    } else {
      tablinks[i].classList.remove("active");
    }
  }
  tablinks = Array.from(pageDIV.querySelectorAll(".tabcontent"));
  console.log(tablinks)
  for (let i = 0; i < tablinks.length; i++) {
    if (tablinks[i] && tablinks[i].getAttribute('id') === tabID) {
      tablinks[i].classList.remove("hidden-tab");
    } else {
      tablinks[i].classList.add("hidden-tab");
    }
  }
}

function correctD2Graph(svg){
  svg.setAttribute('width', "100%");
  svg.setAttribute('height', "100%");
  svg.parentElement.setAttribute('width', "100%");
  svg.parentElement.setAttribute('height', "100%");
  svg.parentElement.setAttribute('preserveAspectRatio', "YMin meet");
}

function getCookie(name) {
  let cookieArr = document.cookie.split("; ");

  for(let i = 0; i < cookieArr.length; i++) {
      let cookiePair = cookieArr[i].split("=");

      if(name == cookiePair[0]) {
          return decodeURIComponent(cookiePair[1]);
      }
  }
  return "";
}

let charts = {};

function updateGraph(pool, data){
  keys = Object.keys(data);
  values = Object.values(data);

  data = {
    labels: keys,
    datasets: [{
      data: values
    }]
  }
  for (const key in charts["ch"+pool]) {
    if (charts["ch"+pool].hasOwnProperty(key)) {
        const ch = charts["ch"+pool][key]
        const element = document.getElementById(`dynamic-chart-${key}`)
        const fontSizeValue = getComputedStyle(element.parentNode).getPropertyValue('font-size')
        const fontSizeInPixels = parseFloat(fontSizeValue)

        if (ch.options.scales) {
          if (ch.options.scales.x && ch.options.scales.x.ticks && ch.options.scales.x.ticks.font) {
            ch.options.scales.x.ticks.font.size = fontSizeInPixels;
          }
          if (ch.options.scales.x && ch.options.scales.x.ticks) {
            ch.options.scales.x.ticks.fontSize = fontSizeInPixels;
          }
          if (ch.options.scales.y && ch.options.scales.y.ticks) {
            ch.options.scales.y.ticks.fontSize = fontSizeInPixels;
          }
          if (ch.options.plugins && ch.options.plugins.legend && ch.options.plugins.legend.labels && ch.options.plugins.legend.labels.font) {
            ch.options.plugins.legend.labels.font.size = fontSizeInPixels;
          }
          if (ch.options.plugins && ch.options.plugins.legend && ch.options.plugins.legend.labels) {
            ch.options.plugins.legend.labels.fontSize = fontSizeInPixels;
          }
        }

        ch.options.fontSize = fontSizeInPixels
        ch.options.font.size = fontSizeInPixels
        ch.data = data
        ch.update()
    }
  }
}

function triggerBlockRun(id) {
  //document.getElementById(id).click();
  castTerminal(id-1)
}

function playAudioIfPaused(id, force) {
  const audio = document.getElementById('audio-page-' + id);
  if (!audio) {
    return
  }
  if (pagesWithAutoPlayAudioPlayed.get("audio-page-"+id)){
    return
  }
  if (audio.currentTime > 0 || force) {
    if (force){
      setTimeout(() => {
        playAudio(id, false);
      }, 1500);
    } else {
      pauseAudio(id);
    }
  }
}

function playAudio(id, user) {
  const audio = document.getElementById('audio-page-' + id);
  if (!audio) {
    return
  }
  if (pagesWithAutoPlayAudioPlayed.get("audio-page-"+id) && !user){
    return
  }
  if (user) {
    pagesWithAutoPlayAudioPlayed.set("audio-page-"+id, false)
  } else {
    pagesWithAutoPlayAudioPlayed.set("audio-page-"+id, true)
  }

  const playBtn = document.getElementById('audio-play-' + id);
  const pauseBtn = document.getElementById('audio-pause-' + id)
  playBtn.style.display = 'none';
  pauseBtn.style.display = 'inline';

  audio.play().catch(error => {
    console.log('ERROR: User interaction required first to play audio.');
  })
}

function pauseAudio(id) {
  const audio = document.getElementById('audio-page-' + id);
  if (!audio) {
    return
  }
  const playBtn = document.getElementById('audio-play-' + id);
  const pauseBtn = document.getElementById('audio-pause-' + id)

  audio.pause();
  pauseBtn.style.display = 'none';
  playBtn.style.display = 'inline';
  console.log(id)
  console.log(audio.currentTime)
}

function resetAudio(id) {
  const audio = document.getElementById('audio-page-' + id)
  if (!audio) {
    return
  }
  const playBtn = document.getElementById('audio-play-' + id)
  const pauseBtn = document.getElementById('audio-pause-' + id)
  audio.pause()
  audio.currentTime = 0
  pauseBtn.style.display = 'none'
  playBtn.style.display = 'inline'
}

  // Optional: reset buttons when audio ends, this needs improvement
  // audio.addEventListener('ended', () => {
  //   pauseBtn.style.display = 'none';
  //   playBtn.style.display = 'inline';
  // });
