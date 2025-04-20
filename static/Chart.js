const charts = {};

export function initChart(hostname) {
  const ctx = document.getElementById('chart-' + hostname);
  if (!ctx) return;

  charts[hostname] = new Chart(ctx, {
    type: 'line',
    data: {
      labels: [],
      datasets: [{
        label: 'CPU %',
        data: [],
        borderColor: '#3b82f6',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
        tension: 0.3,
        borderWidth: 1.5,
        fill: true,
        pointRadius: 0
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: { display: false },
        y: { 
          display: false,
          min: 0,
          max: 100
        }
      },
      plugins: {
        legend: { display: false },
        tooltip: { enabled: false }
      },
      animation: {
        duration: 0
      }
    }
  });
}

export function updateChart(hostname, value) {
  const chart = charts[hostname];
  if (!chart) return;

  const data = chart.data.datasets[0].data;
  const labels = chart.data.labels;

  data.push(value);
  labels.push('');

  if (data.length > 15) {
    data.shift();
    labels.shift();
  }

  chart.update();
}
