import { useEffect, useState } from "react";
import {
  Alert,
  Benchmark,
  Component,
  Deployment,
  Model,
  Page,
  Overview,
  Pod,
  PortalConfig,
} from "./types";
import { api, tone } from "./utils";
import OverviewPage from "./pages/OverviewPage";
import DeploymentsPage from "./pages/DeploymentsPage";
import ModelsPage from "./pages/ModelsPage";
import ObservabilityPage from "./pages/ObservabilityPage";
import AlertsPage from "./pages/AlertsPage";
import LogsPage from "./pages/LogsPage";
import BenchmarkPage from "./pages/BenchmarkPage";
import ConfigPage from "./pages/ConfigPage";

const pages: { id: Page; label: string; icon: string }[] = [
  { id: "overview", label: "Overview", icon: "◈" },
  { id: "deployments", label: "Deployments", icon: "⬡" },
  { id: "models", label: "Models", icon: "◇" },
  { id: "observability", label: "Observability", icon: "◉" },
  { id: "alerts", label: "Alerts", icon: "△" },
  { id: "logs", label: "Logs", icon: "≡" },
  { id: "benchmark", label: "Benchmarks", icon: "↗" },
  { id: "config", label: "Configuration", icon: "⚙" },
];

const emptyOverview: Overview = {
  cluster: "Checking…",
  namespace: "observability",
  gateway: "Unknown",
  observability: "Unknown",
  models: 0,
  alerts: 0,
  pods: 0,
  last_updated: "",
  version: "dev",
  actions_enabled: false,
};

export default function App() {
  const [page, setPage] = useState<Page>(
    () => (location.hash.slice(1) as Page) || "overview",
  );
  const [overview, setOverview] = useState<Overview>(emptyOverview);
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [models, setModels] = useState<Model[]>([]);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [components, setComponents] = useState<Component[]>([]);
  const [pods, setPods] = useState<Pod[]>([]);
  const [config, setConfig] = useState<PortalConfig>({
    namespace: "observability",
    provider: "ollama",
    model: "llama3",
  });
  const [benchmark, setBenchmark] = useState<Benchmark | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const load = async () => {
    setLoading(true);
    setError("");
    try {
      const [o, d, m, a, c, p, cfg, b] = await Promise.all([
        api<Overview>("overview"),
        api<Deployment[]>("deployments"),
        api<Model[]>("models"),
        api<Alert[]>("alerts"),
        api<Component[]>("observability"),
        api<Pod[]>("pods"),
        api<PortalConfig>("config"),
        api<Benchmark | null>("benchmark"),
      ]);
      setOverview(o);
      setDeployments(d);
      setModels(m);
      setAlerts(a);
      setComponents(c);
      setPods(p);
      setConfig(cfg);
      setBenchmark(b);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unable to load portal data");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  useEffect(() => {
    location.hash = page;
  }, [page]);

  const action = async (path: string, body: unknown, message: string) => {
    setNotice("Working…");
    setError("");
    try {
      await api(path, { method: "POST", body: JSON.stringify(body) });
      setNotice(message);
      await load();
    } catch (e) {
      setNotice("");
      setError(e instanceof Error ? e.message : "Action failed");
    }
  };

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">CI</div>
          <div>
            <h1>CloudInferOps</h1>
            <p>Inference control plane</p>
          </div>
        </div>
        <nav>
          {pages.map((item) => (
            <button
              key={item.id}
              className={page === item.id ? "active" : ""}
              onClick={() => setPage(item.id)}
            >
              <span>{item.icon}</span>
              {item.label}
              {item.id === "alerts" && overview.alerts > 0 ? (
                <b>{overview.alerts}</b>
              ) : null}
            </button>
          ))}
        </nav>
        <div className="sidebar-foot">
          <span className={`dot ${tone(overview.cluster)}`} />
          <div>
            <strong>{overview.cluster}</strong>
            <small>
              {overview.namespace} · {overview.version}
            </small>
          </div>
        </div>
      </aside>
      <main>
        <header>
          <div>
            <p className="eyebrow">Workspace / {page}</p>
            <h2>{pages.find((item) => item.id === page)?.label}</h2>
          </div>
          <button
            className="secondary"
            onClick={() => void load()}
            disabled={loading}
          >
            {loading ? "Refreshing…" : "Refresh data"}
          </button>
        </header>
        {error ? (
          <div className="banner error">
            <strong>Something needs attention</strong>
            <span>{error}</span>
            <button onClick={() => setError("")}>×</button>
          </div>
        ) : null}
        {notice ? (
          <div className="banner success">
            <span>{notice}</span>
            <button onClick={() => setNotice("")}>×</button>
          </div>
        ) : null}
        {page === "overview" ? (
          <OverviewPage
            overview={overview}
            deployments={deployments}
            alerts={alerts}
            onNavigate={setPage}
          />
        ) : null}
        {page === "deployments" ? (
          <DeploymentsPage
            items={deployments}
            actions={overview.actions_enabled}
            onDeploy={(target) =>
              void action(
                "actions/deploy",
                { name: target },
                `${target} reconciliation completed`,
              )
            }
            onUndeploy={(target) =>
              void action(
                "actions/undeploy",
                { name: target },
                `${target} uninstallation completed`,
              )
            }
            onRestart={(namespace, name) =>
              void action(
                "actions/restart",
                { namespace, name },
                `Restart requested for ${name}`,
              )
            }
          />
        ) : null}
        {page === "models" ? <ModelsPage items={models} /> : null}
        {page === "observability" ? (
          <ObservabilityPage items={components} />
        ) : null}
        {page === "alerts" ? <AlertsPage items={alerts} /> : null}
        {page === "logs" ? <LogsPage pods={pods} /> : null}
        {page === "benchmark" ? (
          <BenchmarkPage
            result={benchmark}
            actions={overview.actions_enabled}
            model={config.model}
            onRun={async (model, requests) => {
              setNotice("Benchmark is running…");
              try {
                const result = await api<Benchmark>("benchmark", {
                  method: "POST",
                  body: JSON.stringify({ model, requests }),
                });
                setBenchmark(result);
                setNotice("Benchmark completed");
              } catch (e) {
                setNotice("");
                setError(e instanceof Error ? e.message : "Benchmark failed");
              }
            }}
          />
        ) : null}
        {page === "config" ? (
          <ConfigPage
            value={config}
            actions={overview.actions_enabled}
            onSave={async (value) => {
              try {
                await api("config", {
                  method: "PUT",
                  body: JSON.stringify(value),
                });
                setConfig(value);
                setNotice("Configuration saved");
              } catch (e) {
                setError(e instanceof Error ? e.message : "Save failed");
              }
            }}
          />
        ) : null}
      </main>
    </div>
  );
}
