export type Page =
  | "overview"
  | "deployments"
  | "models"
  | "observability"
  | "alerts"
  | "logs"
  | "benchmark"
  | "config";

export interface Overview {
  cluster: string;
  namespace: string;
  gateway: string;
  observability: string;
  models: number;
  alerts: number;
  pods: number;
  last_updated: string;
  version: string;
  actions_enabled: boolean;
}

export interface Deployment {
  name: string;
  namespace: string;
  status: string;
  replicas: string;
  last_updated: string;
}

export interface Model {
  name: string;
  provider: string;
  status: string;
  location: string;
}

export interface Alert {
  title: string;
  severity: string;
  status: string;
  timestamp: string;
}

export interface Component {
  name: string;
  status: string;
  namespace: string;
  url?: string;
}

export interface Pod {
  name: string;
  namespace: string;
  status: string;
}

export interface PortalConfig {
  namespace: string;
  provider: string;
  model: string;
}

export interface Benchmark {
  model: string;
  requests: number;
  succeeded: number;
  failed: number;
  average_latency_ms: number;
  completed_at: string;
}
