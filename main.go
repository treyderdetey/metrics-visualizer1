package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// ServerMetrics модель данных метрик
type ServerMetrics struct {
	Timestamp   int64   `json:"timestamp"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskIO      float64 `json:"disk_io"`
	NetworkIn   float64 `json:"network_in"`
	NetworkOut  float64 `json:"network_out"`
	Uptime      int64   `json:"uptime"`
	Hostname    string  `json:"hostname"`
}

type ServerState struct {
	Uptime      int       `json:"uptime"`
	Status      string    `json:"status"` // "normal", "overloaded", "offline"
	LastChanged time.Time `json:"last_changed"`
}

var (
	clients      = make(map[*websocket.Conn]bool)
	broadcast    = make(chan ServerMetrics)
	mutex        = &sync.Mutex{}
	serversState = make(map[string]*ServerState)
	stateMutex   = &sync.Mutex{}
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// В продакшене замените на конкретный домен
			return true
		},
	}
	oauthConf *oauth2.Config
	rdb       *redis.Client
	ctx       = context.Background()
)

func init() {
	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		log.Print("Note: .env file not found")
	}

	// Инициализация OAuth конфигурации
	oauthConf = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	// Инициализация Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // Default DB
	})
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	mutex.Lock()
	clients[ws] = true
	mutex.Unlock()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, ws)
			mutex.Unlock()
			break
		}
	}
}

func handleMessages() {
	for metric := range broadcast {
		jsonMetric, err := json.Marshal(metric)
		if err != nil {
			log.Printf("JSON marshal error: %v", err)
			continue
		}

		mutex.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, jsonMetric)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

func generateMetricsForServer(hostname string) {
	stateMutex.Lock()
	state, ok := serversState[hostname]
	if !ok {
		status := "normal"
		if hostname == "server-1" || hostname == "server-2" {
			status = "overloaded"
		} else if hostname == "server-3" {
			status = "offline"
		}

		state = &ServerState{
			Uptime:      0,
			Status:      status,
			LastChanged: time.Now(),
		}
		serversState[hostname] = state
	}
	stateMutex.Unlock()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if time.Since(state.LastChanged) > time.Duration(30+rand.Intn(30))*time.Second {
			stateMutex.Lock()
			if hostname == "server-1" || hostname == "server-2" {
				if rand.Float32() < 0.8 {
					state.Status = "overloaded"
				} else {
					state.Status = "normal"
				}
			} else if hostname == "server-3" {
				if rand.Float32() < 0.1 {
					state.Status = "normal"
				} else {
					state.Status = "offline"
				}
			}
			state.LastChanged = time.Now()
			stateMutex.Unlock()
		}

		var metric ServerMetrics
		stateMutex.Lock()
		switch state.Status {
		case "overloaded":
			metric = ServerMetrics{
				Timestamp:   time.Now().Unix(),
				CPUUsage:    80 + rand.Float64()*20,
				MemoryUsage: 75 + rand.Float64()*25,
				DiskIO:      400 + rand.Float64()*600,
				NetworkIn:   150 + rand.Float64()*350,
				NetworkOut:  150 + rand.Float64()*350,
				Uptime:      int64(state.Uptime),
				Hostname:    hostname,
			}
		case "offline":
			state.Uptime++
			stateMutex.Unlock()
			continue
		default:
			metric = ServerMetrics{
				Timestamp:   time.Now().Unix(),
				CPUUsage:    20 + rand.Float64()*40,
				MemoryUsage: 30 + rand.Float64()*40,
				DiskIO:      50 + rand.Float64()*200,
				NetworkIn:   20 + rand.Float64()*80,
				NetworkOut:  20 + rand.Float64()*80,
				Uptime:      int64(state.Uptime),
				Hostname:    hostname,
			}
		}
		state.Uptime++
		stateMutex.Unlock()

		broadcast <- metric
	}
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	url := oauthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	token, err := oauthConf.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := oauthConf.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userinfo struct {
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userinfo); err != nil {
		http.Error(w, "Failed to decode user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sessionID := fmt.Sprintf("sess-%d", time.Now().UnixNano())
	if err := rdb.Set(ctx, sessionID, userinfo.Picture, 24*time.Hour).Err(); err != nil {
		http.Error(w, "Failed to save session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	http.Redirect(w, r, "/dashboard.html", http.StatusSeeOther)
}

func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Unauthorized: missing session", http.StatusUnauthorized)
		return
	}

	avatar, err := rdb.Get(ctx, cookie.Value).Result()
	if err != nil {
		http.Error(w, "Unauthorized: invalid session", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"avatar": avatar})
}

func main() {
	// Проверка подключения к Redis
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	// Запуск генерации метрик
	hostnames := []string{"server-1", "server-2", "server-3", "server-4", "server-5"}
	for _, hostname := range hostnames {
		go generateMetricsForServer(hostname)
	}

	// Обработчик сообщений
	go handleMessages()

	// HTTP роуты
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/auth", handleAuth)
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/api/userinfo", handleUserInfo)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.Handle("/dashboard", http.RedirectHandler("/dashboard.html", http.StatusSeeOther))

	// Настройка graceful shutdown
	server := &http.Server{Addr: ":8080"}
	done := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Закрытие WebSocket соединений
		mutex.Lock()
		for client := range clients {
			client.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""))
			client.Close()
		}
		mutex.Unlock()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown error: %v", err)
		}
		close(done)
	}()

	log.Println("Server starting on :8080")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe error: %v", err)
	}
	<-done
	log.Println("Server stopped gracefully")
}
