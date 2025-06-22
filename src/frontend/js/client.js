let socket;
let currentUser;
let currentChat = { type: "public", with: "" };
let typingTimeout;
let lastTypingSent = 0;

window.onload = () => {
  const saved = localStorage.getItem("gochat_username");
  if (saved) {
    currentUser = saved;
    document.getElementById("auth-modal").style.display = "none";
    document.getElementById("chat-container").classList.remove("hidden");
    startChat();
  }

  document.getElementById("auth-action").onclick = handleAuth;
  document.getElementById("toggle-auth").onclick = toggleAuthMode;
  document.getElementById("global-chat-btn").onclick = () => selectChat("");
};

async function handleAuth() {
  const mode = document.getElementById("auth-action").textContent.toLowerCase();
  const user = document.getElementById("auth-username").value.trim();
  const pass = document.getElementById("auth-password").value;
  if (!user || !pass) return;

  const res = await fetch(`/${mode}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username: user, password_hash: pass }),
  });

  if (res.ok) {
    currentUser = user;
    localStorage.setItem("gochat_username", user);
    document.getElementById("auth-modal").style.display = "none";
    document.getElementById("chat-container").classList.remove("hidden");
    startChat();
  } else {
    document.getElementById("auth-error").textContent = await res.text();
  }
}

function toggleAuthMode() {
  const action = document.getElementById("auth-action");
  const title = document.getElementById("auth-title");
  const toggle = document.getElementById("toggle-auth");
  const label = document.getElementById("auth-switch").firstChild;

  if (action.textContent === "Register") {
    action.textContent = "Login";
    title.textContent = "Login";
    toggle.textContent = "Register";
    label.textContent = "Don’t have an account? ";
  } else {
    action.textContent = "Register";
    title.textContent = "Register";
    toggle.textContent = "Login";
    label.textContent = "Already have an account? ";
  }
  document.getElementById("auth-error").textContent = "";
}

function startChat() {
  setupSocket();
  loadUsers();
  bindChatUI();
  selectChat("");
}

function bindChatUI() {
  const input = document.getElementById("msg-input");
  const btn = document.getElementById("send-btn");

  input.disabled = false;
  btn.disabled = false;

  btn.onclick = sendMessage;
  input.onkeydown = (e) => {
    if (e.key === "Enter") sendMessage();
  };
  input.oninput = sendTyping;
}

function setupSocket() {
  socket = new WebSocket(`ws://${location.host}/ws`);
  socket.onopen = () => {
    socket.send(currentUser);
    requestHistory();
  };
  socket.onmessage = (e) => handleIncoming(JSON.parse(e.data));
}

function handleIncoming(msg) {
  if (msg.type === "typing") {
    const inThis =
      (msg.receiver === "" && currentChat.type === "public") ||
      (msg.receiver === currentUser && msg.sender === currentChat.with);
    if (inThis) showTyping(msg.sender);
    return;
  }

  if (msg.type === "history" && Array.isArray(msg.messages)) {
    msg.messages.forEach((m) => {
      if (m.text && m.text.trim()) appendMessage(m.sender, m.text);
    });
    return;
  }

  const isPublic = msg.receiver === "" || msg.receiver === undefined;
  const shouldShow =
    (isPublic && currentChat.type === "public") ||
    (msg.receiver === currentUser && msg.sender === currentChat.with) ||
    (msg.receiver === currentChat.with && msg.sender === currentUser);

  if (shouldShow) appendMessage(msg.sender, msg.text);
}

function appendMessage(sender, text) {
  const div = document.createElement("div");
  div.textContent = `${sender}: ${text}`;
  const box = document.getElementById("chat-box");
  box.appendChild(div);
  box.scrollTop = box.scrollHeight;
}

function sendMessage() {
  const input = document.getElementById("msg-input");
  const text = input.value.trim();
  if (!text || socket.readyState !== WebSocket.OPEN) return;

  const msg = {
    type: "message",
    sender: currentUser,
    receiver: currentChat.type === "dm" ? currentChat.with : "",
    text: text,
  };

  socket.send(JSON.stringify(msg));
  input.value = "";
}

function sendTyping() {
  const now = Date.now();
  if (now - lastTypingSent < 200 || socket.readyState !== WebSocket.OPEN)
    return;
  lastTypingSent = now;

  socket.send(
    JSON.stringify({
      type: "typing",
      sender: currentUser,
      receiver: currentChat.type === "dm" ? currentChat.with : "",
    }),
  );
}

function showTyping(name) {
  const elm = document.getElementById("typingIndicator");
  elm.textContent = `${name} is typing...`;
  clearTimeout(typingTimeout);
  typingTimeout = setTimeout(() => {
    elm.textContent = "";
  }, 1000);
}

function selectChat(username) {
  currentChat = username
    ? { type: "dm", with: username }
    : { type: "public", with: "" };

  document.querySelectorAll(".user-list li").forEach((li) => {
    li.classList.toggle("active", li.textContent === username);
  });
  clearChat();
  requestHistory();
}

function clearChat() {
  document.getElementById("chat-box").innerHTML = "";
}

function requestHistory() {
  if (socket.readyState !== WebSocket.OPEN) return;
  socket.send(
    JSON.stringify({
      type: "load_history",
      sender: currentUser,
      receiver: currentChat.type === "dm" ? currentChat.with : "",
    }),
  );
}

async function loadUsers() {
  const res = await fetch("/users");
  const names = await res.json();
  const ul = document.getElementById("users");
  ul.innerHTML = "";
  names
    .filter((n) => n !== currentUser)
    .forEach((n) => {
      const li = document.createElement("li");
      li.textContent = n;
      li.onclick = () => selectChat(n);
      ul.appendChild(li);
    });
}
