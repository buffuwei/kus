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
                // term.write(text);
                ws.send(text);
              })
        };
        return true;
    });

    return true;
});

const fitAddon =  new FitAddon()
term.loadAddon(fitAddon);
term.open(document.getElementById("terminal"));
term.focus();
// 监听键盘输入
term.onKey(function (e) {
    ws.send(e.key);
});

fitAddon.fit();
// 监听窗口大小变化事件
window.addEventListener("resize", function () {
    fitAddon.fit();
});

let clu = document.getElementById("cluster").innerText.trim();
let ns = document.getElementById("ns").innerText.trim();
let pod = document.getElementById("pod").innerText.trim();
let container = document.getElementById("container").innerText.trim();
let url = "ws://localhost:9900/ws?cluster="+clu+"&ns="+ns+"&pod="+pod+"&container="+container;
console.log(url);

const ws = new WebSocket(url);
ws.onopen = function () {
    console.log("WebSocket connection opened");
};

ws.onmessage = function (event) {
    if (event.data === "0") {
        // ignore
    } else {
        term.write(event.data);
    }
};

ws.onerror = function (error) {
    console.log("WebSocket error:", error);
};

ws.onclose = function () {
    console.log("WebSocket connection closed");
};

