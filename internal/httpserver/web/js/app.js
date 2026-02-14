import { renderLogin } from './auth.js';
import { renderChat } from './chat.js';

document.addEventListener('DOMContentLoaded', () => {
  const token = localStorage.getItem('token');
  if (token) {
    renderChat(); // Uses token to fetch /whoami
  } else {
    renderLogin();
  }
});
