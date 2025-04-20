<template>
    <div class="p-4 rounded-2xl border shadow transition-all"
         :class="statusColor">
      <h2 class="font-bold text-lg mb-1">{{ name }}</h2>
      <p class="text-gray-500 text-sm">{{ ip }}</p>
  
      <div v-for="(val, label) in metrics" :key="label" class="mt-3">
        <p class="text-sm font-medium">{{ label }}</p>
        <div class="w-full bg-gray-200 rounded-full h-2 mt-1">
          <div class="h-2 rounded-full"
               :style="{ width: val + '%' }"
               :class="progressColor(label)">
          </div>
        </div>
        <div class="text-right text-xs text-gray-500">{{ val }}%</div>
      </div>
    </div>
  </template>
  
  <script setup lang="ts">
  const props = defineProps<{
    name: string,
    ip: string,
    status: string,
    metrics: Record<string, number>
  }>()
  
  const statusColor = computed(() => {
    return props.status === 'High Load'
      ? 'border-yellow-500'
      : props.status === 'Offline'
      ? 'border-red-500'
      : 'border-green-500'
  })
  
  function progressColor(label: string) {
    if (label === 'CPU') return 'bg-blue-500'
    if (label === 'RAM') return 'bg-purple-500'
    if (label === 'Диск') return 'bg-yellow-400'
    return 'bg-gray-400'
  }
  </script>
  