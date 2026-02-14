import { renderChat } from './chat.js';

export function renderLogin() {
  const app = document.getElementById('app');
  app.innerHTML = `
    <div class="auth-container">
      <div class="auth-card">
        <h1>LI-CHAT 🤖</h1>
        <div class="error-message" id="errorMsg"></div>
        <input id="username" placeholder="Username" required />
        <input id="password" type="password" placeholder="Password" required />
        <div class="button-group">
          <button onclick="renderRegister()">Register</button>
          <button id="loginBtn">Login</button>
        </div>
      </div>
    </div>
  `;

  document.getElementById('loginBtn').addEventListener('click', handleLogin);
  document.getElementById('password').addEventListener('keypress', e => {
    if (e.key === 'Enter') handleLogin();
  });
}

export function renderRegister() {
  const app = document.getElementById('app');
  app.innerHTML = `
    <div class="auth-container">
      <div class="auth-card">
        <h1>Create Account</h1>
        <div class="error-message" id="errorMsg"></div>
        <input id="username" placeholder="Username" required />
        <input id="password" type="password" placeholder="Password" required />
        <div class="button-group">
          <button onclick="renderLogin()">Back</button>
          <button id="registerBtn">Register</button>
        </div>
      </div>
    </div>
  `;

  document.getElementById('registerBtn').addEventListener('click', handleRegister);
}

async function handleLogin() {
  const username = document.getElementById('username').value.trim();
  const password = document.getElementById('password').value.trim();
  if (!username || !password) return showError("Enter username and password");

  const loginBtn = document.getElementById('loginBtn');
  loginBtn.disabled = true;
  loginBtn.textContent = "Logging in...";

  try {
    const res = await fetch('/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });

    const text = await res.text();
    let data;
    try { data = JSON.parse(text); } catch { throw new Error(text); }

    if (!res.ok) throw new Error(data.message || 'Login failed');

    localStorage.setItem('token', data.token);
    renderChat();
  } catch (err) {
    showError(err.message || "Network error");
    loginBtn.disabled = false;
    loginBtn.textContent = "Login";
  }
}

async function handleRegister() {
  const username = document.getElementById('username').value.trim();
  const password = document.getElementById('password').value.trim();
  if (!username || !password) return showError("Enter username and password");

  const btn = document.getElementById('registerBtn');
  btn.disabled = true;
  btn.textContent = "Creating...";

  try {
    const res = await fetch('/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });

    const text = await res.text();
    let data;
    try { data = JSON.parse(text); } catch { throw new Error(text); }

    if (!res.ok) throw new Error(data.message || 'Registration failed');

    localStorage.setItem('token', data.token || '');
    handleLogin();
  } catch (err) {
    showError(err.message || "Network error");
    btn.disabled = false;
    btn.textContent = "Register";
  }
}

export async function authFetch(url, options = {}) {
  const token = localStorage.getItem('token');
  if (!token) throw new Error("No token");

  const headers = { 'Content-Type': 'application/json', ...(options.headers || {}), 'Authorization': `Bearer ${token}` };

  const res = await fetch(url, { ...options, headers });
  if (res.status === 401) {
    localStorage.removeItem('token');
    renderLogin();
    throw new Error("Session expired");
  }
  return res;
}

function showError(message) {
  const el = document.getElementById('errorMsg');
  if (el) {
    el.textContent = message;
    el.classList.add('show');
    setTimeout(() => el.classList.remove('show'), 5000);
  }
}
