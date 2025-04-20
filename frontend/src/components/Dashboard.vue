<template>
    <div class="p-6">
      <h1 class="text-2xl font-bold mb-4">Мониторинг серверов</h1>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <ServerCard
          v-for="server in servers"
          :key="server.name"
          :name="server.name"
          :ip="server.ip"
          :status="server.status"
          :metrics="server.metrics"
        />
      </div>
    </div>
  </template>
  
  <script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import ServerCard from '../components/ServerCard.vue'
  import type { ServerMetric } from '../types/metrics'
  
  const servers = ref<ServerMetric[]>([])
  
  onMounted(() => {
    const ws = new WebSocket('ws://localhost:8080/ws')
  
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data)
      const newMetric: ServerMetric = {
        name: data.name,
        ip: data.ip || '192.168.0.1',
        status: data.cpu > 85 ? 'High Load' : 'Online',
        metrics: {
          CPU: Math.round(data.cpu),
          RAM: Math.round(data.memory),
          Диск: Math.round(data.disk),
        },
      }
  
      const index = servers.value.findIndex(s => s.name === newMetric.name)
      if (index !== -1) {
        servers.value[index] = newMetric
      } else {
        servers.value.push(newMetric)
      }
    }
  })
  </script>
  