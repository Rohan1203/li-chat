import { login, register, logout } from "./auth.js";
import { loadMessages, initChat, sendMessage } from "./chat.js";
import { apiFetch } from "./api.js";

let currentUser = null;

export function getCurrentUser() {
  return currentUser;
}

export function renderLogin() {
  const app = document.getElementById("app");

  app.innerHTML = `
    <div class="auth-card">
      <h2>GoChat</h2>
      <input id="username" placeholder="Username"/>
      <input id="password" type="password" placeholder="Password"/>
      <button id="loginBtn">Login</button>
      <button id="registerBtn" class="secondary">Register</button>
      <p id="error" class="error"></p>
    </div>
  `;

  document.getElementById("loginBtn").onclick = async () => {
    try {
      await login(
        username.value.trim(),
        password.value.trim()
      );
    } catch (err) {
      error.textContent = err.message;
    }
  };

  document.getElementById("registerBtn").onclick = async () => {
    try {
      await register(
        username.value.trim(),
        password.value.trim()
      );
    } catch (err) {
      error.textContent = err.message;
    }
  };
}

export async function renderChat() {
  const app = document.getElementById("app");
  console.log("renderChat called");

  const res = await apiFetch("/whoami");
  const data = await res.json();
  currentUser = data.username;

  app.innerHTML = `
    <div class="chat-layout">
      <header>
        <div>GoChat</div>
        <div>${escapeHtml(currentUser)}</div>
        <button id="logoutBtn">Logout</button>
      </header>

      <main id="messages"></main>

      <footer>
        <input id="messageInput" placeholder="Type message..." />
        <button id="sendBtn">Send</button>
      </footer>

      <div id="status" class="status offline">Offline</div>
    </div>
  `;

  document.getElementById("logoutBtn").onclick = logout;

  document.getElementById("sendBtn").onclick = () => {
    const input = messageInput;
    const text = input.value.trim();
    if (!text) return;
    sendMessage(text);
    input.value = "";
  };

  await loadMessages(renderMessage);

  initChat(renderMessage, updateStatus);
}

function renderMessage(msg) {
  const container = document.getElementById("messages");
  const div = document.createElement("div");
  div.className = msg.username === currentUser ? "msg own" : "msg";

  div.innerHTML = `
    <span class="user">${escapeHtml(msg.username)}</span>
    <span class="text">${escapeHtml(msg.content)}</span>
  `;

  container.appendChild(div);
  container.scrollTop = container.scrollHeight;
}

function updateStatus(online) {
  const el = document.getElementById("status");
  el.textContent = online ? "Online" : "Reconnecting...";
  el.className = online ? "status online" : "status offline";
}

function escapeHtml(str) {
  const div = document.createElement("div");
  div.textContent = str;
  return div.innerHTML;
}
