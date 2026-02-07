// Chat UI and WebSocket logic
let ws = null;
let currentUser = null;
let messageBuffer = [];

async function renderChat() {
  const app = document.getElementById('app');
  
  // Get current user
  try {
    const token = localStorage.getItem('token');
    if (!token) {
      renderLogin();
      return;
    }

    const res = await fetch('/whoami', {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    if (res.ok) {
      const data = await res.json();
      currentUser = data.username;
    } else {
      localStorage.removeItem('token');
      renderLogin();
      return;
    }
  } catch (err) {
    console.error('Error getting user info:', err);
    renderLogin();
    return;
  }
  
  app.innerHTML = `
    <div class="chat-container">
      <div class="chat-header">
        <div class="chat-title">💬 Go Chat</div>
        <div class="chat-user">Welcome, <strong>${escapeHtml(currentUser)}</strong>!</div>
        <button class="logout-btn" onclick="handleLogout()">Logout</button>
      </div>
      
      <div class="chat-main">
        <div class="messages-container" id="messagesContainer">
          <div class="empty-state">
            <p>Welcome to Go Chat! 👋</p>
            <p style="font-size: 12px; margin-top: 10px;">Start a conversation</p>
          </div>
        </div>
        
        <div class="input-container">
          <input 
            type="text" 
            class="message-input" 
            id="messageInput" 
            placeholder="Type a message..."
            autocomplete="off"
          >
          <button class="send-btn" id="sendBtn" onclick="sendMessage()">Send</button>
        </div>
      </div>
      
      <div class="connection-status disconnected" id="connectionStatus">
        <span class="status-indicator"></span>
        <span>Connecting...</span>
      </div>
    </div>
  `;
  
  // Setup event listeners
  const messageInput = document.getElementById('messageInput');
  const messagesContainer = document.getElementById('messagesContainer');
  
  messageInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  });
  
  // Auto-scroll to bottom when new messages arrive
  const observer = new MutationObserver(() => {
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
  });
  observer.observe(messagesContainer, { childList: true });
  
  // Connect to WebSocket
  connectWebSocket();
}

function connectWebSocket() {
  const token = localStorage.getItem('token');
  if (!token) {
    renderLogin();
    return;
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/ws?token=${encodeURIComponent(token)}`;
  
  try {
    ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
      console.log('WebSocket connected');
      updateConnectionStatus(true);
      
      // Send buffered messages
      messageBuffer.forEach(msg => {
        ws.send(JSON.stringify(msg));
      });
      messageBuffer = [];
    };
    
    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        displayMessage(message);
      } catch (err) {
        console.error('Error parsing message:', err);
      }
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      updateConnectionStatus(false);
    };
    
    ws.onclose = () => {
      console.log('WebSocket closed');
      updateConnectionStatus(false);
      
      // Attempt to reconnect after 3 seconds
      setTimeout(() => {
        console.log('Attempting to reconnect...');
        connectWebSocket();
      }, 3000);
    };
  } catch (err) {
    console.error('Error connecting to WebSocket:', err);
    updateConnectionStatus(false);
  }
}

function sendMessage() {
  const input = document.getElementById('messageInput');
  const message = input.value.trim();
  
  if (!message) return;
  
  const msgData = {
    username: currentUser,
    content: message
  };
  
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(msgData));
  } else {
    // Buffer message if not connected
    messageBuffer.push(msgData);
    updateConnectionStatus(false);
  }
  
  input.value = '';
  input.focus();
}

function displayMessage(message) {
  const container = document.getElementById('messagesContainer');
  
  // Clear empty state
  const emptyState = container.querySelector('.empty-state');
  if (emptyState) {
    emptyState.remove();
  }
  
  const isOwn = message.username === currentUser;
  const messageDiv = document.createElement('div');
  messageDiv.className = `message ${isOwn ? 'message-own' : 'message-other'}`;
  
  const bubble = document.createElement('div');
  bubble.className = 'message-bubble';
  bubble.textContent = message.content;
  
  const meta = document.createElement('div');
  meta.className = 'message-meta';
  meta.innerHTML = `<strong>${escapeHtml(message.username)}</strong> • ${formatTime(message.created_at)}`;
  
  messageDiv.appendChild(bubble);
  messageDiv.appendChild(meta);
  container.appendChild(messageDiv);
  
  // Scroll to bottom
  container.scrollTop = container.scrollHeight;
}

function formatTime(timestamp) {
  try {
    const date = new Date(timestamp);
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${hours}:${minutes}`;
  } catch (err) {
    return timestamp;
  }
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

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}
