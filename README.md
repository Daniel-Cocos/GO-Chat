# GO-Chat

A simple, real-time web chat application built with Go for the backend and vanilla JavaScript, HTML, and CSS for the frontend. GO-Chat uses WebSockets to enable instant messaging between connected clients.

## Features

- Message encryption
- WebSocket-based real-time messaging  
- Go backend with Gorilla WebSocket  
- Chat history (locally stored and encrypted)
- Private/BroadCast messaging (client-to-client)

## Project Structure

```
GO-Chat
├── data
│   ├── dm
│   │   └── User_User1.dm
│   ├── public_chat.history
│   └── users.json
├── go.mod
├── go.sum
├── LICENSE
├── README.md
└── src
    ├── backend
    │   ├── decrypt.py
    │   ├── encrypt.py
    │   ├── main.go
    │   └── password_validation.py
    └── frontend
        ├── css
        │   └── style.css
        ├── index.html
        └── js
            └── client.js
```
- The dm directory will contain as many ecrypted conversations between users as there are

## Getting Started

1. Clone the repository:
    ```bash
    git clone https://github.com/Daniel-Cocos/GO-Chat.git
    cd GO-Chat
    ```

2. Run the server:
    ```bash
    go run src/backend/main.go
    ```

3. Open browser and navigate to:
    ```
    http://localhost:8080
    ```

4. Optional if trying it with two different devices on the same network:
    ```bash
    ip a | grep inet # on linux / macOS
    ipconfig    # on windows
    ```

    ```
    https://192.168.yourip:8080
    ```

## How It Works

- The Go backend upgrades HTTP connections to WebSockets.
- Connected clients can send messages using the UI.
- Each message is broadcast to all other connected clients.
- Messages are dynamically added to the chat box.
