<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import type { ServerMetrics } from '../types/metrics'; // Создадим этот файл позже

// Определяем тип для хранения последних метрик по hostname
const latestMetrics = ref<Record<string, ServerMetrics>>({});

// Переменная для WebSocket соединения
let websocket: WebSocket | null = null;
const connectionStatus = ref('Connecting...');

// URL вашего бэкенда (напрямую, пока без Nginx/Docker Compose)
const websocketUrl = 'ws://localhost:8080/ws';

// Функция для подключения к WebSocket
const connectWebSocket = () => {
  connectionStatus.value = 'Connecting...';
  websocket = new WebSocket(websocketUrl);

  // Обработчик открытия соединения
  websocket.onopen = () => {
    connectionStatus.value = 'Connected';
    console.log('WebSocket connection opened');
  };

  // Обработчик получения сообщения
  websocket.onmessage = (event) => {
    try {
      const metric: ServerMetrics = JSON.parse(event.data);
      // Обновляем метрики для соответствующего hostname
      latestMetrics.value = {
        ...latestMetrics.value, // Копируем существующие метрики
        [metric.hostname]: metric // Обновляем или добавляем метрики для текущего хоста
      };
      // console.log('Received metric:', metric); // Для отладки
    } catch (error) {
      console.error('Error parsing message:', error);
    }
  };

  // Обработчик ошибки
  websocket.onerror = (event) => {
    connectionStatus.value = 'Error';
    console.error('WebSocket error:', event);
    if (websocket) {
        websocket.close(); // Закрываем соединение при ошибке
    }
  };

  // Обработчик закрытия соединения
  websocket.onclose = (event) => {
    connectionStatus.value = 'Disconnected';
    console.log('WebSocket connection closed:', event);
    // Попытка переподключиться через некоторое время (опционально, но хорошо для устойчивости)
    // setTimeout(connectWebSocket, 5000); // Например, через 5 секунд
  };
};


// Хук жизненного цикла Vue: выполняется после монтирования компонента
onMounted(() => {
  connectWebSocket(); // Подключаемся при загрузке компонента
});

// Хук жизненного цикла Vue: выполняется перед размонтированием компонента
onUnmounted(() => {
  if (websocket) {
    websocket.close(); // Закрываем соединение при уходе со страницы
  }
});

// Вспомогательная функция для форматирования чисел
const formatValue = (value: number | undefined): string => {
    if (value === undefined || value === null) return 'N/A';
    return value.toFixed(2);
};
 const formatUptime = (value: number | undefined): string => {
    if (value === undefined || value === null) return 'N/A';
    // Простая конвертация в минуты/секунды для удобства
    const minutes = Math.floor(value / 60);
    const seconds = value % 60;
    return `${minutes}m ${seconds}s`;
};

</script>

<template>
  <div>
    <h2>Server Metrics Dashboard</h2>
    <p>Connection Status: {{ connectionStatus }}</p>

    <div class="metrics-list">
      <!-- Итерируемся по объекту latestMetrics, ключи - хостнеймы, значения - метрики -->
      <div v-for="(metric, hostname) in latestMetrics" :key="hostname" class="server-card">
        <h3>{{ hostname }}</h3>
        <p><strong>Timestamp:</strong> {{ metric.timestamp }}</p>
        <p><strong>CPU Usage:</strong> {{ formatValue(metric.cpu_usage) }} %</p>
        <p><strong>Memory Usage:</strong> {{ formatValue(metric.memory_usage) }} %</p>
        <p><strong>Disk I/O:</strong> {{ formatValue(metric.disk_io) }} MB/s</p>
        <p><strong>Network In:</strong> {{ formatValue(metric.network_in) }} MB/s</p>
        <p><strong>Network Out:</strong> {{ formatValue(metric.network_out) }} MB/s</p>
        <p><strong>Uptime:</strong> {{ formatUptime(metric.uptime) }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.metrics-list {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.server-card {
  border: 1px solid #ccc;
  padding: 15px;
  border-radius: 8px;
  width: 300px;
  box-shadow: 2px 2px 8px rgba(0,0,0,0.1);
  background-color: #f9f9f9;
  /* Можно также добавить цвет текста здесь, если хотите, чтобы все внутри карточки было темным */
  /* color: #333; */
}

.server-card h3 {
    margin-top: 0;
    color: #333; /* Заголовок уже темный */
}

.server-card p {
    margin: 5px 0;
    font-size: 0.9em;
    /* !!! ДОБАВЬТЕ ЭТУ СТРОКУ: !!! */
    color: #333; /* Устанавливаем темный цвет текста для параграфов */
}

 .server-card strong {
    display: inline-block;
    min-width: 100px;
}
</style>