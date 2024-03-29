import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import './style.css'

console.log("Hello kus-term!");

const term = new Terminal({
    cursorBlink: true,
    cursorStyle: 'block',
    scrollback: 1000,
    fontSize: 13,
    fontFamily: 'Consolas,Menlo,Ubuntu Mono,Bitstream Vera Sans Mono,Monaco,"微软雅黑",monospace',
    fontWeight: 549,
    theme: {
        foreground: '#E6EDF4',
        background: '#202124',
        cursor: 'help',
        selection: 'rgba(0, 0, 0, 0.3)',
    }
});

term.attachCustomKeyEventHandler(function (event) {
    term.attachCustomKeyEventHandler((arg) => {
        if (arg.metaKey && arg.code === "KeyV" && arg.type === "keydown") {
            navigator.clipboard.readText()
              .then(text => {
                let msg = "0" + btoa(text)
                ws.send(msg);
              })
        };
        return true;
    });

    return true;
});

const fitAddon =  new FitAddon()
fitAddon.activate(term)

let containerEle = document.getElementById("terminal")
term.open(containerEle);
term.focus();

term.onKey(function (e) {
    let msg = "0" + btoa(e.key)
    ws.send(msg);
});

term.onResize(function (e) {
    let s = JSON.stringify({"Width":e.cols,"Height":e.rows})
    let msg = "4" + btoa(s);
    ws.send(msg)
})

window.addEventListener("resize", function () {
    console.log("window resize event listened")
    fitAddon.fit();
});

let pageURL = new URL(window.location.href)
let clu = document.getElementById("cluster").innerText.trim();
let ns = document.getElementById("ns").innerText.trim();
let pod = document.getElementById("pod").innerText.trim();
let container = document.getElementById("container").innerText.trim();
let url = "ws://localhost:"+pageURL.port+"/ws?cluster="+clu+"&ns="+ns+"&pod="+pod+"&container="+container;
console.log(url);

const ws = new WebSocket(url);
ws.onopen = function () {
    console.log("WebSocket connection opened");
};

var fristMsgReceived = false;

ws.onmessage = function (event) {
    if (event.data === "0") {
        // ignore
    } else {
        if (!fristMsgReceived) {
            console.log("First msg", event.data)
            fristMsgReceived = true;
            fitAddon.fit();
        }
        term.write(event.data);
    }
};

ws.onerror = function (error) {
    console.log("WebSocket error:", error);
};

ws.onclose = function () {
    console.log("WebSocket connection closed");
};

