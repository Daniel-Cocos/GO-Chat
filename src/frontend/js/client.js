// Open WebSocket connection to the GO server
const socket = new WebSocket(`ws://${location.host}/ws`);

socket.onopen = function () {
  console.log("Connected to WebSocket server");
};

// Wait for messages sent from server
socket.onmessage = function (event) {
  const chatBox = document.getElementById("chat-box");
  const message = document.createElement("div");
  message.textContent = "Server: " + event.data;
  chatBox.appendChild(message);
};

socket.onerror = function (error) {
  console.error("WebSocket Error:", error);
};

// Send message into socket
function sendMessage() {
  const input = document.getElementById("msg-input");
  const msg = input.value;
  if (msg.trim() === "") return;

  const chatBox = document.getElementById("chat-box");
  const message = document.createElement("div");
  message.textContent = "You: " + msg;
  chatBox.appendChild(message);

  socket.send(msg); // Send to server
  input.value = ""; // Clear input
}

// Enter key sends messages
const inputField = document.getElementById("msg-input");
inputField.addEventListener('keydown', function(event) {
  if (event.key === 'Enter') {
    sendMessage();
    event.preventDefault();
  }
});
