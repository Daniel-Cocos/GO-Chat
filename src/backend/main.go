package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

type Message struct {
	Sender    string    `json:"sender"`
	Receiver  string    `json:"receiver"`
	Text      string    `json:"text"`
	Timestamp string    `json:"timestamp,omitempty"`
	Type      string    `json:"type"`
	Messages  []Message `json:"messages,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var (
	clients     = make(map[*websocket.Conn]string)
	conns       = make(map[string]*websocket.Conn)
	broadcast   = make(chan Message)
	historyLock sync.Mutex
	connLock    sync.Mutex
)

func main() {
	setupRoutes()
	go handleMessages()
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupRoutes() {
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/users", handleUsers)
	http.Handle("/", http.FileServer(http.Dir("src/frontend")))
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	var u User
	json.NewDecoder(r.Body).Decode(&u)
	u.Username = strings.TrimSpace(u.Username)
	if u.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("python3", "src/backend/password_validation.py", u.PasswordHash)
	if out, err := cmd.CombinedOutput(); err != nil {
		http.Error(w, string(out), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Hash error", http.StatusInternalServerError)
		return
	}
	u.PasswordHash = string(hash)

	users := readUsers()
	for _, user := range users {
		if user.Username == u.Username {
			http.Error(w, "Username taken", http.StatusConflict)
			return
		}
	}
	users = append(users, u)
	saveUsers(users)
	w.WriteHeader(http.StatusCreated)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	var input User
	json.NewDecoder(r.Body).Decode(&input)
	users := readUsers()
	for _, user := range users {
		if user.Username == input.Username {
			if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.PasswordHash)) == nil {
				w.WriteHeader(http.StatusOK)
				return
			}
			break
		}
	}
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	users := readUsers()
	names := []string{}
	for _, u := range users {
		names = append(names, u.Username)
	}
	json.NewEncoder(w).Encode(names)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}
	username := string(msg)

	connLock.Lock()
	clients[conn] = username
	conns[username] = conn
	connLock.Unlock()

	defer func() {
		connLock.Lock()
		delete(clients, conn)
		delete(conns, username)
		connLock.Unlock()
		conn.Close()
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var m Message
		json.Unmarshal(raw, &m)
		m.Timestamp = time.Now().Format(time.RFC3339)

		if m.Type == "load_history" {
			go sendHistory(conn, m.Sender, m.Receiver)
			continue
		}

		broadcast <- m
	}
}

func handleMessages() {
	for msg := range broadcast {
		saveMessage(msg)
		data, _ := json.Marshal(msg)

		if msg.Receiver == "" {
			for conn := range clients {
				conn.WriteMessage(websocket.TextMessage, data)
			}
		} else {
			if c, ok := conns[msg.Receiver]; ok {
				c.WriteMessage(websocket.TextMessage, data)
			}
			if c, ok := conns[msg.Sender]; ok {
				c.WriteMessage(websocket.TextMessage, data)
			}
		}
	}
}

func sendHistory(conn *websocket.Conn, sender, receiver string) {
	path := "data/public_chat.history"
	if receiver != "" {
		pair := []string{sender, receiver}
		sort.Strings(pair)
		path = fmt.Sprintf("data/dm/%s_%s.dm", pair[0], pair[1])
	}
	historyLock.Lock()
	defer historyLock.Unlock()

	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var messages []Message
	for scanner.Scan() {
		line := scanner.Text()
		plain := decryptWithPython(line)
		parts := strings.SplitN(plain, "|", 3)
		if len(parts) == 3 {
			messages = append(messages, Message{
				Sender:    parts[0],
				Timestamp: parts[1],
				Text:      parts[2],
			})
		}
	}

	resp := Message{
		Type:     "history",
		Messages: messages,
	}
	data, _ := json.Marshal(resp)
	conn.WriteMessage(websocket.TextMessage, data)
}

func readUsers() []User {
	f, err := os.ReadFile("data/users.json")
	if err != nil {
		return nil
	}
	var users []User
	json.Unmarshal(f, &users)
	return users
}

func saveUsers(users []User) {
	os.MkdirAll("data", 0755)
	out, _ := json.MarshalIndent(users, "", "  ")
	os.WriteFile("data/users.json", out, 0644)
}

func saveMessage(msg Message) {
	key := "public_chat.history"
	if msg.Receiver != "" {
		pair := []string{msg.Sender, msg.Receiver}
		sort.Strings(pair)
		key = fmt.Sprintf("dm/%s_%s.dm", pair[0], pair[1])
		os.MkdirAll("data/dm", 0755)
	}
	plaintext := fmt.Sprintf("%s|%s|%s", msg.Sender, msg.Timestamp, msg.Text)
	enc := encryptWithPython(plaintext)

	historyLock.Lock()
	defer historyLock.Unlock()
	f, _ := os.OpenFile("data/"+key, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(enc + "\n")
}

func encryptWithPython(text string) string {
	cmd := exec.Command("python3", "src/backend/encrypt.py", text)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return text
	}
	return strings.TrimSpace(string(out))
}

func decryptWithPython(text string) string {
	cmd := exec.Command("python3", "src/backend/decrypt.py", text)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return text
	}
	return strings.TrimSpace(string(out))
}
