// Файл: metrics-visualizer/internal/model/metric.go
package model

// Импортируем стандартный пакет Go для работы со временем

// ServerMetrics представляет собой набор метрик для одного сервера в определенный момент времени.
// Поля структуры соответствуют JSON-формату, который мы будем использовать.
// Теги `json:"..."` указывают Go, как сопоставить поля структуры с ключами в JSON.
type ServerMetrics struct {
	Timestamp   int64   `json:"timestamp"`    // Время генерации метрики (Unix Timestamp)
	CPUUsage    float64 `json:"cpu_usage"`    // Использование CPU в процентах (например, 64.7)
	MemoryUsage float64 `json:"memory_usage"` // Использование RAM в процентах (например, 48.2)
	DiskIO      float64 `json:"disk_io"`      // Скорость Disk I/O в МБ/с (например, 732.5)
	NetworkIn   float64 `json:"network_in"`   // Скорость входящего сетевого трафика в МБ/с (например, 129.7)
	NetworkOut  float64 `json:"network_out"`  // Скорость исходящего сетевого трафика в МБ/с (например, 117.9)
	Uptime      int64   `json:"uptime"`       // Время работы сервера в секундах (например, 1023)
	Hostname    string  `json:"hostname"`     // Имя сервера (например, "server-1")
}

// Примечание: В твоем JSON timestamp указан как число.
// В предыдущих примерах мы использовали time.Time, но чтобы точно соответствовать JSON,
// используем int64 (для Unix Timestamp). При генерации нужно будет получать текущее время
// и преобразовывать его в int64.
