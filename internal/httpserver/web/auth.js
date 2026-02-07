// Authentication UI and logic
function renderLogin() {
  const app = document.getElementById('app');
  
  app.innerHTML = `
    <div class="auth-container">
      <div class="auth-card">
        <div class="auth-header">
          <h1>💬 Go Chat</h1>
          <p>Real-time messaging made simple</p>
        </div>
        
        <div class="error-message" id="errorMsg"></div>
        <div class="success-message" id="successMsg"></div>
        
        <form id="authForm">
          <div class="form-group">
            <label for="username">Username</label>
            <input 
              type="text" 
              id="username" 
              placeholder="Enter your username"
              required
              autocomplete="username"
            >
          </div>
          
          <div class="form-group">
            <label for="password">Password</label>
            <input 
              type="password" 
              id="password" 
              placeholder="Enter your password"
              required
              autocomplete="current-password"
            >
          </div>
          
          <div class="button-group">
            <button type="button" class="btn btn-secondary" id="registerBtn" onclick="switchToRegister()">
              Register
            </button>
            <button type="button" class="btn btn-primary" id="loginBtn" onclick="handleLogin()">
              Login
            </button>
          </div>
        </form>
        
        <div style="text-align: center; margin-top: 20px; color: #999; font-size: 12px;">
          <p>Create a new account or login to start chatting</p>
        </div>
      </div>
    </div>
  `;
  
  // Add event listeners
  const form = document.getElementById('authForm');
  form.addEventListener('submit', (e) => {
    e.preventDefault();
    handleLogin();
  });
  
  // Allow Enter key to submit
  document.getElementById('password').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
      handleLogin();
    }
  });
}

async function handleLogin() {
  const username = document.getElementById('username').value.trim();
  const password = document.getElementById('password').value.trim();
  const errorMsg = document.getElementById('errorMsg');
  const loginBtn = document.getElementById('loginBtn');
  
  // Clear messages
  errorMsg.classList.remove('show');
  errorMsg.textContent = '';
  
  if (!username || !password) {
    showError('Please enter both username and password');
    return;
  }
  
  try {
    loginBtn.disabled = true;
    loginBtn.textContent = '...';
    
    const res = await fetch('/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });
    
    if (res.ok) {
      const data = await res.json();
      // Store JWT token in localStorage
      localStorage.setItem('token', data.token);
      // Login successful, render chat
      renderChat();
    } else {
      const text = await res.text();
      showError(text || 'Login failed. Please check your credentials.');
      loginBtn.disabled = false;
      loginBtn.textContent = 'Login';
    }
  } catch (err) {
    console.error('Login error:', err);
    showError('Network error. Please try again.');
    loginBtn.disabled = false;
    loginBtn.textContent = 'Login';
  }
}

function switchToRegister() {
  const username = document.getElementById('username').value;
  const password = document.getElementById('password').value;
  renderRegister(username, password);
}

function renderRegister(username = '', password = '') {
  const app = document.getElementById('app');
  
  app.innerHTML = `
    <div class="auth-container">
      <div class="auth-card">
        <div class="auth-header">
          <h1>💬 Go Chat</h1>
          <p>Create your account</p>
        </div>
        
        <div class="error-message" id="errorMsg"></div>
        <div class="success-message" id="successMsg"></div>
        
        <form id="authForm">
          <div class="form-group">
            <label for="username">Username</label>
            <input 
              type="text" 
              id="username" 
              placeholder="Choose a username"
              value="${username}"
              required
              autocomplete="username"
            >
          </div>
          
          <div class="form-group">
            <label for="password">Password</label>
            <input 
              type="password" 
              id="password" 
              placeholder="Create a password"
              value="${password}"
              required
              autocomplete="new-password"
            >
          </div>
          
          <div class="button-group">
            <button type="button" class="btn btn-secondary" id="backBtn" onclick="renderLogin()">
              Back
            </button>
            <button type="button" class="btn btn-primary" id="registerBtn" onclick="handleRegister()">
              Register
            </button>
          </div>
        </form>
        
        <div style="text-align: center; margin-top: 20px; color: #999; font-size: 12px;">
          <p>Register to create a new account</p>
        </div>
      </div>
    </div>
  `;
  
  // Add event listeners
  document.getElementById('password').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
      handleRegister();
    }
  });
}

async function handleRegister() {
  const username = document.getElementById('username').value.trim();
  const password = document.getElementById('password').value.trim();
  const errorMsg = document.getElementById('errorMsg');
  const successMsg = document.getElementById('successMsg');
  const registerBtn = document.getElementById('registerBtn');
  
  // Clear messages
  errorMsg.classList.remove('show');
  successMsg.classList.remove('show');
  errorMsg.textContent = '';
  successMsg.textContent = '';
  
  if (!username || !password) {
    showError('Please enter both username and password');
    return;
  }
  
  if (password.length < 3) {
    showError('Password must be at least 3 characters');
    return;
  }
  
  try {
    registerBtn.disabled = true;
    registerBtn.textContent = '...';
    
    const res = await fetch('/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });
    
    if (res.ok) {
      // Registration successful
      successMsg.textContent = 'Account created! Logging in...';
      successMsg.classList.add('show');
      
      // Auto-login after registration
      setTimeout(() => {
        document.getElementById('username').value = username;
        document.getElementById('password').value = password;
        handleLogin();
      }, 1000);
    } else {
      const text = await res.text();
      showError(text || 'Registration failed. Username may already exist.');
      registerBtn.disabled = false;
      registerBtn.textContent = 'Register';
    }
  } catch (err) {
    console.error('Register error:', err);
    showError('Network error. Please try again.');
    registerBtn.disabled = false;
    registerBtn.textContent = 'Register';
  }
}

function showError(message) {
  const errorMsg = document.getElementById('errorMsg');
  errorMsg.textContent = message;
  errorMsg.classList.add('show');
  
  // Auto hide after 5 seconds
  setTimeout(() => {
    errorMsg.classList.remove('show');
  }, 5000);
}

async function handleLogout() {
  try {
    const token = localStorage.getItem('token');
    await fetch('/logout', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` }
    });
  } catch (err) {
    console.error('Logout error:', err);
  }
  
  // Remove token and return to login
  localStorage.removeItem('token');
  renderLogin();
}
