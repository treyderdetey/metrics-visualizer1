export interface ServerMetrics {
    timestamp: number;
    cpu_usage: number;
    memory_usage: number;
    disk_io: number;
    network_in: number;
    network_out: number;
    uptime: number;
    hostname: string;
  }