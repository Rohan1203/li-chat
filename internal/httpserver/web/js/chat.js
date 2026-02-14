import { authFetch } from './auth.js';

let ws = null;
let currentUser = null;
let messageBuffer = [];

export async function renderChat() {
  const app = document.getElementById('app');

  // Get current user
  try {
    const res = await authFetch('/whoami');
    const text = await res.text();
    let data;
    try { data = JSON.parse(text); } catch { throw new Error(text); }
    currentUser = data.username;
  } catch (err) {
    console.error("Failed to get user:", err);
    localStorage.removeItem('token');
    location.reload();
    return;
  }
  let seconds = 0;
  const clockEl = document.getElementById("liveClock");

    function updateClock() {
    const clockEl = document.getElementById("liveClock");
    if (!clockEl) return; // Prevent crash if not found

    const now = new Date();

    const timeString = now.toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hour12: false
    });

    clockEl.textContent = timeString;
    }

    // Run once
    updateClock();

    // Update every second
    setInterval(updateClock, 1000);

  app.innerHTML = `
    <div class="chat-container">
      <div class="chat-header">
        <div class="chat-title">LI-CHAT 🤖</div>
        <div class="chat-user">Welcome, <strong>${escapeHtml(currentUser)}</strong>
            <span id="liveClock" style="margin-left:12px; font-size:0.85rem; opacity:0.85;"></span>
        </div>
        
        <button class="logout-btn" id="logoutBtn">Logout</button>
      </div>
      <div class="chat-main">
        <div class="messages-container" id="messagesContainer">
          <div class="empty-state"><p>Loading messages...</p></div>
        </div>
        <div class="input-container">
         <label class="code-toggle">
            <input type="checkbox" id="codeMode" />
            Code
        </label>
          <textarea 
            id="messageInput" 
            placeholder="Type a message..." 
            rows="1"></textarea>
          <button id="sendBtn">Send</button>
        </div>
      </div>
      <div class="connection-status disconnected" id="connectionStatus">
        <span class="status-indicator"></span>
        <span>Connecting...</span>
      </div>
    </div>
  `;

  document.getElementById('logoutBtn').addEventListener('click', handleLogout);
  document.getElementById('sendBtn').addEventListener('click', sendMessage);
  document.getElementById('messageInput').addEventListener('keypress', e => {
    if (e.key === 'Enter') sendMessage();
  });

  await loadMessageHistory();
  connectWebSocket();
}

// Load old messages
async function loadMessageHistory() {
  try {
    const res = await authFetch('/messages');
    const text = await res.text();
    let messages;
    try { messages = JSON.parse(text); } catch { messages = []; }

    const container = document.getElementById('messagesContainer');
    container.innerHTML = '';

    if (!messages || messages.length === 0) {
      container.innerHTML = `<div class="empty-state"><p>Welcome to Go Chat! 👋</p></div>`;
      return;
    }

    messages.forEach(displayMessage);
  } catch (err) {
    console.error("Load messages failed:", err);
  }
}

function connectWebSocket() {
  const token = localStorage.getItem('token');
  if (!token) return;

  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${location.host}/ws?token=${encodeURIComponent(token)}`;

  ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    console.log("WebSocket connected");
    updateConnectionStatus(true);
    messageBuffer.forEach(msg => ws.send(JSON.stringify(msg)));
    messageBuffer = [];
  };

  ws.onmessage = e => {
    try { displayMessage(JSON.parse(e.data)); } catch {}
  };

  ws.onclose = () => {
    console.log("WebSocket closed. Reconnecting in 3s...");
    updateConnectionStatus(false);
    setTimeout(connectWebSocket, 3000);
  };

  ws.onerror = err => {
    console.error("WS error:", err);
    ws.close();
  };
}

function sendMessage() {
  const input = document.getElementById('messageInput');
  const codeCheckbox = document.getElementById('codeMode');

  let message = input.value.trim();
  if (!message) return;

  // Wrap in triple backticks if checkbox is checked
  if (codeCheckbox && codeCheckbox.checked) {
    message = `\`\`\`\n${message}\n\`\`\``;
  }

  const msgData = { username: currentUser, content: message };

  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(msgData));
  } else {
    messageBuffer.push(msgData);
    updateConnectionStatus(false);
  }

  // Clear input
  input.value = '';
  input.style.height = 'auto';

  // Optional: auto-uncheck after send
  if (codeCheckbox) codeCheckbox.checked = false;
}


function displayMessage(message) {
  const container = document.getElementById('messagesContainer');
  const emptyState = container.querySelector('.empty-state');
  if (emptyState) emptyState.remove();

  const isOwn = message.username === currentUser;

  // Static avatars
  const avatars = {
    "Alice": "./images/avatar1.png",
    "Bob": "./images/avatar2.png",
  };
  const avatar = avatars[message.username] || "./assets/images/avatar-default.png";

  // Format timestamp
  const time = new Date(message.timestamp || Date.now());
  const hours = time.getHours().toString().padStart(2, '0');
  const minutes = time.getMinutes().toString().padStart(2, '0');
  const formattedTime = `${hours}:${minutes}`;

  const div = document.createElement('div');
  div.className = `message ${isOwn ? 'message-own' : 'message-other'}`;
  div.innerHTML = `
    <img src="${avatar}" alt="${escapeHtml(message.username)}" class="message-avatar" />
    <div class="message-content">
      <div class="message-bubble">${formatMessage(message.content)}</div>
      <div class="message-meta">
        <strong>${escapeHtml(message.username)}</strong>
        <span class="timestamp">${formattedTime}</span>
      </div>
    </div>
  `;

  container.appendChild(div);
  container.scrollTop = container.scrollHeight;
}

function updateConnectionStatus(connected) {
  const status = document.getElementById('connectionStatus');
  if (!status) return;

  if (connected) {
    status.classList.remove('disconnected');
    status.classList.add('connected');
    status.innerHTML = '<span class="status-indicator"></span><span>Connected</span>';
  } else {
    status.classList.remove('connected');
    status.classList.add('disconnected');
    status.innerHTML = '<span class="status-indicator"></span><span>Disconnected</span>';
  }
}

async function handleLogout() {
  try { await authFetch('/logout', { method: 'POST' }); } catch (_) {}
  localStorage.removeItem('token');
  location.reload();
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// code block
function formatMessage(text) {
  const escaped = escapeHtml(text);

  return escaped.replace(/```(\w+)?\n?([\s\S]*?)```/g, (match, lang, code) => {
    return `
<pre class="code-block"><code class="language-${lang || 'plain'}">${code.replace(/^\n+|\n+$/g, '')}</code></pre>
    `;
  });
}