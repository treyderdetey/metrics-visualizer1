package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"metrics-visualizer/internal/model"

	"github.com/gorilla/websocket"
)

type ServerState struct {
	Uptime      int
	Status      string // "normal", "overloaded", "offline"
	LastChanged time.Time
}

var (
	clients      = make(map[*websocket.Conn]bool)
	broadcast    = make(chan model.ServerMetrics)
	mutex        = &sync.Mutex{}
	serversState = make(map[string]*ServerState)
	stateMutex   = &sync.Mutex{}
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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

func generateMetricsForServer(hostname string, metricChan chan<- model.ServerMetrics) {
	// Инициализация состояния сервера
	stateMutex.Lock()
	state, ok := serversState[hostname]
	if !ok {
		// Определяем начальный статус сервера
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
		log.Printf("Initialized server %s with status: %s", hostname, status)
	}
	stateMutex.Unlock()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Периодически меняем статусы (каждые 30-60 секунд)
		if time.Since(state.LastChanged) > time.Duration(30+rand.Intn(30))*time.Second {
			stateMutex.Lock()
			// Случайным образом меняем статус, но сохраняем общую логику
			if hostname == "server-1" || hostname == "server-2" {
				// 80% chance to stay overloaded, 20% to become normal
				if rand.Float32() < 0.8 {
					state.Status = "overloaded"
				} else {
					state.Status = "normal"
				}
			} else if hostname == "server-3" {
				// 90% chance to stay offline, 10% to become normal
				if rand.Float32() < 0.1 {
					state.Status = "normal"
				} else {
					state.Status = "offline"
				}
			} else {
				// 10% chance to change status for normal servers
				if rand.Float32() < 0.1 {
					if rand.Float32() < 0.3 {
						state.Status = "overloaded"
					} else {
						state.Status = "normal"
					}
				}
			}
			state.LastChanged = time.Now()
			stateMutex.Unlock()
		}

		var metric model.ServerMetrics
		stateMutex.Lock()
		switch state.Status {
		case "overloaded":
			// Генерация перегруженных метрик
			metric = model.ServerMetrics{
				Timestamp:   time.Now().Unix(),
				CPUUsage:    80 + rand.Float64()*20,   // 80-100%
				MemoryUsage: 75 + rand.Float64()*25,   // 75-100%
				DiskIO:      400 + rand.Float64()*600, // 400-1000 MB/s
				NetworkIn:   150 + rand.Float64()*350, // 150-500 MB/s
				NetworkOut:  150 + rand.Float64()*350, // 150-500 MB/s
				Uptime:      int64(state.Uptime),
				Hostname:    hostname,
			}
		case "offline":
			// Сервер offline - не отправляем метрики
			state.Uptime++ // Увеличиваем uptime, но не отправляем данные
			stateMutex.Unlock()
			continue
		default:
			// Нормальная работа
			metric = model.ServerMetrics{
				Timestamp:   time.Now().Unix(),
				CPUUsage:    20 + rand.Float64()*40,  // 20-60%
				MemoryUsage: 30 + rand.Float64()*40,  // 30-70%
				DiskIO:      50 + rand.Float64()*200, // 50-250 MB/s
				NetworkIn:   20 + rand.Float64()*80,  // 20-100 MB/s
				NetworkOut:  20 + rand.Float64()*80,  // 20-100 MB/s
				Uptime:      int64(state.Uptime),
				Hostname:    hostname,
			}
		}
		state.Uptime++
		stateMutex.Unlock()

		metricChan <- metric
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	hostnames := []string{
		"server-1", // Перегружен
		"server-2", // Перегружен
		"server-3", // Отключен
		"server-4", // Нормальный
		"server-5", // Нормальный
	}

	go handleMessages()

	for _, hostname := range hostnames {
		go generateMetricsForServer(hostname, broadcast)
	}

	http.HandleFunc("/ws", handleConnections)
	log.Println("Server starting on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server error: ", err)
	}
}
