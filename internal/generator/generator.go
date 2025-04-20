// Файл: metrics-visualizer/internal/generator/generator.go
package generator

import (
	"context"   // Для обработки сигналов завершения (graceful shutdown)
	"log"       // Для вывода информации и ошибок
	"math/rand" // Для генерации случайных чисел
	"sync"      // Для синхронизации горутин (WaitGroup)
	"time"      // Для работы со временем (тикеры, Unix Timestamp)

	"metrics-visualizer/internal/model" // Импортируем нашу структуру ServerMetrics
)

// Config содержит настройки для генератора
type Config struct {
	ServerIDs []string      // Список имен серверов для симуляции
	Interval  time.Duration // Интервал генерации метрик для каждого сервера
}

// Generator отвечает за генерацию синтетических метрик
type Generator struct {
	config      Config                     // Конфигурация генератора
	output      chan<- model.ServerMetrics // Канал, куда отправлять сгенерированные метрики (только для записи)
	serverState map[string]*ServerUptime   // Состояние каждого сервера (здесь только uptime)
	mu          sync.Mutex                 // Мьютекс для защиты serverState (если понадобится)
}

// ServerUptime хранит состояние для отдельного симулируемого сервера
type ServerUptime struct {
	Uptime int64 // Время работы в секундах
}

// New создает новый экземпляр Генератора
// Принимает конфигурацию и канал, куда будут отправляться метрики
func New(cfg Config, output chan<- model.ServerMetrics) *Generator {
	// Инициализируем состояние для каждого сервера
	state := make(map[string]*ServerUptime)
	for _, id := range cfg.ServerIDs {
		state[id] = &ServerUptime{Uptime: 0} // Начинаем с 0 секунд uptime
	}

	// Важно: Инициализировать генератор случайных чисел один раз
	// В более сложных случаях можно использовать crypto/rand для безопасности,
	// но для синтетических данных math/rand достаточно.
	// Используем time.Now().UnixNano() как зерно для большей случайности при каждом запуске
	rand.Seed(time.Now().UnixNano())

	return &Generator{
		config:      cfg,
		output:      output,
		serverState: state,
	}
}

// Start запускает процесс генерации метрик.
// Эта функция запускает отдельную горутину для каждого симулируемого сервера.
// Она блокируется до тех пор, пока не будет получен сигнал завершения через контекст.
func (g *Generator) Start(ctx context.Context, wg *sync.WaitGroup) {
	log.Println("Starting data generator...")

	// Запускаем отдельную горутину для каждого симулируемого сервера
	for _, serverID := range g.config.ServerIDs {
		// Используем замыкание, чтобы передать правильный serverID в горутину
		id := serverID
		state := g.serverState[id] // Получаем состояние для этого сервера

		wg.Add(1) // Увеличиваем счетчик горутин, за которыми следим
		go func() {
			defer wg.Done()                      // Уменьшаем счетчик, когда горутина завершается
			g.runServerGenerator(ctx, id, state) // Запускаем функцию генерации для сервера
		}()
	}

	// Горутина Start теперь ждет, пока контекст не будет отменен.
	// Это позволяет функции Start немедленно вернуться, а реальная работа
	// происходит в запущенных выше горутинах runServerGenerator.
	// Когда ctx.Done() сработает, это будет сигналом для runServerGenerator завершиться.
	<-ctx.Done()

	log.Println("Data generator received shutdown signal. Waiting for server generators to finish...")
	// WaitGroup wg передается в main, где будет ждать завершения всех горутин.
}

// runServerGenerator генерирует метрики для одного конкретного сервера
// и отправляет их в выходной канал.
func (g *Generator) runServerGenerator(ctx context.Context, serverID string, state *ServerUptime) {
	log.Printf("Starting generator for %s with interval %s", serverID, g.config.Interval)

	// Создаем тикер, который будет срабатывать каждые g.config.Interval
	ticker := time.NewTicker(g.config.Interval)
	defer ticker.Stop() // Обязательно останавливаем тикер при выходе из функции

	for {
		select {
		case <-ticker.C:
			// Тикер сработал, пора генерировать метрику
			metric := g.generateMetric(serverID, state)

			// Отправляем метрику в выходной канал.
			// Используем select с default, чтобы не заблокироваться,
			// если канал переполнен (хотя с буферизованным каналом это маловероятно).
			select {
			case g.output <- metric:
				// log.Printf("Generated and sent metric for %s", serverID) // Слишком много логов
			default:
				// Если канал полный, просто пропускаем эту метрику и выводим предупреждение.
				log.Printf("Warning: Metric channel full, dropping metric for %s", serverID)
			}

		case <-ctx.Done():
			// Получен сигнал завершения через контекст. Пора выходить.
			log.Printf("Generator for %s received context done. Shutting down.", serverID)
			return // Выходим из цикла и завершаем горутину
		}
	}
}

// generateMetric создает одну структуру ServerMetrics с синтетическими данными
func (g *Generator) generateMetric(serverID string, state *ServerUptime) model.ServerMetrics {
	// Генерируем случайные значения в заданных диапазонах
	// Используем rand.Float64() * (макс - мин) + мин для чисел с плавающей точкой
	// Используем rand.Int63n(макс - мин) + мин для целых чисел

	cpuUsage := rand.Float64()*(95.0-5.0) + 5.0
	memoryUsage := rand.Float64()*(80.0-10.0) + 10.0
	diskIO := rand.Float64() * 500.0     // От 0 до 500
	networkIn := rand.Float64() * 200.0  // От 0 до 200
	networkOut := rand.Float64() * 200.0 // От 0 до 200

	// Увеличиваем uptime сервера
	state.Uptime++ // Увеличиваем на 1 каждую секунду

	// Получаем текущий Unix Timestamp (количество секунд с 1 января 1970 UTC)
	timestamp := time.Now().Unix()

	// Создаем и заполняем структуру ServerMetrics
	metric := model.ServerMetrics{
		Timestamp:   timestamp,
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		DiskIO:      diskIO,
		NetworkIn:   networkIn,
		NetworkOut:  networkOut,
		Uptime:      state.Uptime,
		Hostname:    serverID, // Используем переданный ID сервера
	}

	return metric
}
