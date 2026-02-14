import { CONFIG } from "./config.js";

let socket = null;
let reconnectDelay = CONFIG.RECONNECT_BASE_DELAY;

export function connectSocket(onMessage, onStatusChange) {
    console.log("connectSocket called");
  const token = localStorage.getItem("token");
  if (!token) return;

  const protocol = location.protocol === "https:" ? "wss:" : "ws:";
  const url = `${protocol}//${location.host}${CONFIG.WS_PATH}?token=${encodeURIComponent(token)}`;

  socket = new WebSocket(url);

  socket.onopen = () => {
    reconnectDelay = CONFIG.RECONNECT_BASE_DELAY;
    onStatusChange(true);
  };

  socket.onmessage = (event) => {
    onMessage(JSON.parse(event.data));
  };

  socket.onclose = () => {
    onStatusChange(false);
    setTimeout(() => {
      reconnectDelay = Math.min(reconnectDelay * 2, CONFIG.RECONNECT_MAX_DELAY);
      connectSocket(onMessage, onStatusChange);
    }, reconnectDelay);
  };

  socket.onerror = () => socket.close();
}

export function sendSocketMessage(payload) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify(payload));
  }
}
