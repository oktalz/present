var page = /^#?\d+$/.test(window.location.hash) ? parseInt(window.location.hash.slice(1), 10) : 0;
setPage(page);
var spinner = false
var myID = ""
var showMenu = false

function setSpinner(value){
    spinner = value
    if (value) {
        document.getElementById("run-"+page+"-refresh").classList.remove("closed")
        document.getElementById("run-"+page+"").classList.add("closed")
    } else {
        document.getElementById("run-"+page+"-refresh").classList.add("closed")
        document.getElementById("run-"+page+"").classList.remove("closed")
    }
}

function getSlideElements() {
  return document.querySelectorAll('[id^="slide-"]');
}

function setPageWithUpdate(newPage) {
  oldPage = page;
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
  page = newPage
  if (page < 0) {
    page = 0;
  }
  if (page > maxPage) {
    page = maxPage
  }
  window.location.hash = page.toString();
  updateSlideVisibility(page);
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

function updateSlideVisibility(page) {
  getSlideElements().forEach(function(slide) {
    if (parseInt(slide.id.slice(6)) == page) {
      slide.classList.remove('hidden');
    } else {
      slide.classList.add('hidden');
    }
  });
}

function closeTerminal(){
  const terminalElement = document.getElementById('terminal-'+page);
  const terminalElementX = document.getElementById('terminalx-'+page);
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
    oldPage = page;
    page = page + 1;
    target = getPageUp(oldPage,page);
    setPage(target);
    updateData({
      Author: myID,
      Slide: page
    })
  }
  if (previousPageKeys.includes(keyCode)) {
    oldPage = page;
    page = page - 1;
    target = getPageDown(oldPage,page);
    setPage(target);
    updateData({
      Author: myID,
      Slide: page
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
      let index = page
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
      oldPage = page;
      page = page + 1;
      target = getPageUp(oldPage,page);
      setPage(target);
      updateData({
        Author: myID,
        Slide: page
      })
  }
  if (scrollDirection == 'upward') {
      oldPage = page;
      page = page - 1;
      target = getPageDown(oldPage,page);
      setPage(target);
      updateData({
        Author: myID,
        Slide: page
      })
  }
}, true);

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
    oldPage = page;
    page = page - 1;
    target = getPageDown(oldPage,page);
    setPage(target);
    updateData({
      Author: myID,
      Slide: page
    })
  }

  if (endX < touchX) {
    oldPage = page;
    page = page + 1;
    target = getPageUp(oldPage,page);
    setPage(target);
    updateData({
      Author: myID,
      Slide: page
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
  let pageDIV = document.getElementById("slide-"+page);
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
