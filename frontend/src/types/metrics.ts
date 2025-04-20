export interface ServerMetric {
  name: string
  ip: string
  status: string
  metrics: {
    CPU: number
    RAM: number
    Диск: number
  }
}
