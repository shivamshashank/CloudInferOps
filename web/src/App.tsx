import { useEffect, useMemo, useState } from "react";

type Page =
  | "overview"
  | "deployments"
  | "models"
  | "observability"
  | "alerts"
  | "logs"
  | "benchmark"
  | "config";
type Overview = {
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
};
type Deployment = {
  name: string;
  namespace: string;
  status: string;
  replicas: string;
  last_updated: string;
};
type Model = {
  name: string;
  provider: string;
  status: string;
  location: string;
};
type Alert = {
  title: string;
  severity: string;
  status: string;
  timestamp: string;
};
type Component = {
  name: string;
  status: string;
  namespace: string;
  url?: string;
};
type Pod = { name: string; namespace: string; status: string };
type PortalConfig = { namespace: string; provider: string; model: string };
type Benchmark = {
  model: string;
  requests: number;
  succeeded: number;
  failed: number;
  average_latency_ms: number;
  completed_at: string;
};

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

async function api<T>(path: string, options?: RequestInit): Promise<T> {
  const token = sessionStorage.getItem("cloudinferops-token");
  const response = await fetch(`api/${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options?.headers,
    },
  });
  const payload = await response
    .json()
    .catch(() => ({ message: "Invalid server response" }));
  if (!response.ok)
    throw new Error(payload.message || `Request failed (${response.status})`);
  return payload as T;
}
const tone = (status: string) =>
  /running|healthy|connected|active|available|ok/i.test(status)
    ? "good"
    : /partial|pending|warning|unknown|checking/i.test(status)
      ? "warn"
      : "muted";
const when = (value: string) =>
  value ? new Date(value).toLocaleString() : "—";

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

function Status({ value }: { value: string }) {
  return (
    <span className={`status ${tone(value)}`}>
      <i />
      {value}
    </span>
  );
}
function Empty({ title, copy }: { title: string; copy: string }) {
  return (
    <div className="empty">
      <div>◇</div>
      <strong>{title}</strong>
      <p>{copy}</p>
    </div>
  );
}
function OverviewPage({
  overview,
  deployments,
  alerts,
  onNavigate,
}: {
  overview: Overview;
  deployments: Deployment[];
  alerts: Alert[];
  onNavigate: (p: Page) => void;
}) {
  const latest = alerts[0];
  return (
    <>
      <section className="hero">
        <div>
          <span className="live">
            <i /> LIVE CONTROL PLANE
          </span>
          <h3>
            Operate your inference stack
            <br />
            <em>with a clear signal.</em>
          </h3>
          <p>
            One place for cluster health, model serving, telemetry, and safe
            operational actions.
          </p>
        </div>
        <div className="orb">
          <span>{overview.pods}</span>
          <small>pods discovered</small>
        </div>
      </section>
      <section className="metric-grid">
        <article>
          <span>Cluster</span>
          <strong>{overview.cluster}</strong>
          <Status value={overview.namespace} />
        </article>
        <article>
          <span>Gateway</span>
          <strong>{overview.gateway}</strong>
          <small>Inference edge</small>
        </article>
        <article>
          <span>Observability</span>
          <strong>{overview.observability}</strong>
          <small>Telemetry stack</small>
        </article>
        <article>
          <span>Models</span>
          <strong>{overview.models}</strong>
          <small>Available in cluster</small>
        </article>
      </section>
      <section className="two-col">
        <article className="panel">
          <div className="panel-head">
            <div>
              <span className="kicker">RUNTIME</span>
              <h3>Deployment health</h3>
            </div>
            <button className="text" onClick={() => onNavigate("deployments")}>
              View all →
            </button>
          </div>
          {deployments.slice(0, 5).map((d) => (
            <div className="row" key={`${d.namespace}/${d.name}`}>
              <div>
                <strong>{d.name}</strong>
                <small>{d.namespace}</small>
              </div>
              <Status value={d.status} />
              <span>{d.replicas}</span>
            </div>
          ))}
          {deployments.length === 0 ? (
            <Empty
              title="No deployments found"
              copy="Connect a Kubernetes cluster or deploy a workload to begin."
            />
          ) : null}
        </article>
        <article className="panel">
          <div className="panel-head">
            <div>
              <span className="kicker">LATEST SIGNAL</span>
              <h3>Incident activity</h3>
            </div>
            <button className="text" onClick={() => onNavigate("alerts")}>
              Open alerts →
            </button>
          </div>
          {latest ? (
            <div className="incident">
              <Status value={latest.severity} />
              <h4>{latest.title}</h4>
              <p>
                {latest.status} · {when(latest.timestamp)}
              </p>
            </div>
          ) : (
            <Empty
              title="All quiet"
              copy="No active incidents were returned by the webhook gateway."
            />
          )}
          <div className="updated">
            Last synchronized {when(overview.last_updated)}
          </div>
        </article>
      </section>
    </>
  );
}

function DeploymentsPage({
  items,
  actions,
  onDeploy,
  onRestart,
}: {
  items: Deployment[];
  actions: boolean;
  onDeploy: (x: string) => void;
  onRestart: (n: string, x: string) => void;
}) {
  const [query, setQuery] = useState("");
  const filtered = items.filter((i) =>
    `${i.name} ${i.namespace}`.toLowerCase().includes(query.toLowerCase()),
  );
  return (
    <section className="panel">
      <div className="toolbar">
        <input
          aria-label="Search deployments"
          placeholder="Search deployments…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <div className="button-group">
          <button disabled={!actions} onClick={() => onDeploy("inference")}>
            Deploy inference
          </button>
          <button
            className="secondary"
            disabled={!actions}
            onClick={() => onDeploy("observability")}
          >
            Deploy observability
          </button>
        </div>
      </div>
      <p className="guard">
        {actions
          ? "Operational actions are enabled for this installation."
          : "Read-only mode: set CLOUDINFEROPS_UI_ENABLE_ACTIONS=1 to enable guarded actions."}
      </p>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Deployment</th>
              <th>Namespace</th>
              <th>Health</th>
              <th>Ready</th>
              <th>Created</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {filtered.map((item) => (
              <tr key={`${item.namespace}/${item.name}`}>
                <td>
                  <strong>{item.name}</strong>
                </td>
                <td>{item.namespace}</td>
                <td>
                  <Status value={item.status} />
                </td>
                <td>{item.replicas}</td>
                <td>{when(item.last_updated)}</td>
                <td>
                  <button
                    className="tiny"
                    disabled={!actions}
                    onClick={() => onRestart(item.namespace, item.name)}
                  >
                    Restart
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {filtered.length === 0 ? (
        <Empty
          title="No matching deployments"
          copy="Try a different search or deploy the inference stack."
        />
      ) : null}
    </section>
  );
}
function ModelsPage({ items }: { items: Model[] }) {
  const [query, setQuery] = useState("");
  const filtered = useMemo(
    () =>
      items.filter((i) =>
        `${i.name} ${i.provider}`.toLowerCase().includes(query.toLowerCase()),
      ),
    [items, query],
  );
  return (
    <section className="panel">
      <div className="toolbar">
        <input
          aria-label="Search models"
          placeholder="Search model inventory…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <span>{filtered.length} models</span>
      </div>
      <div className="card-grid">
        {filtered.map((item) => (
          <article className="model-card" key={item.name}>
            <div className="model-icon">◇</div>
            <Status value={item.status} />
            <h3>{item.name}</h3>
            <p>{item.provider} provider</p>
            <footer>
              <span>{item.location}</span>
              <span>OpenAI compatible</span>
            </footer>
          </article>
        ))}
      </div>
      {filtered.length === 0 ? (
        <Empty
          title="No models discovered"
          copy="Deploy Ollama or vLLM and make sure the gateway /models endpoint is reachable."
        />
      ) : null}
    </section>
  );
}
function ObservabilityPage({ items }: { items: Component[] }) {
  return (
    <>
      <section className="section-intro">
        <div>
          <h3>Telemetry stack</h3>
          <p>Live discovery across the observability namespace.</p>
        </div>
        <Status
          value={`${items.filter((i) => i.status === "Running").length}/${
            items.length
          } running`}
        />
      </section>
      <section className="card-grid">
        {items.map((item) => (
          <article className="component-card" key={item.name}>
            <div className="component-top">
              <div className="model-icon">◉</div>
              <Status value={item.status} />
            </div>
            <h3>{item.name}</h3>
            <p>{item.namespace}</p>
            <div className="signal-line">
              <i className={tone(item.status)} />
              <span>
                {item.status === "Running"
                  ? "Metrics and health signals available"
                  : "Waiting for a ready deployment"}
              </span>
            </div>
          </article>
        ))}
      </section>
    </>
  );
}
function AlertsPage({ items }: { items: Alert[] }) {
  const [filter, setFilter] = useState("all");
  const filtered = items.filter(
    (i) => filter === "all" || i.status.toLowerCase() === filter,
  );
  return (
    <section className="panel">
      <div className="toolbar">
        <div className="tabs">
          {["all", "active", "resolved"].map((x) => (
            <button
              key={x}
              className={filter === x ? "active" : ""}
              onClick={() => setFilter(x)}
            >
              {x}
            </button>
          ))}
        </div>
        <span>{filtered.length} incidents</span>
      </div>
      {filtered.map((item) => (
        <div className="alert-row" key={`${item.title}/${item.timestamp}`}>
          <div className={`severity ${tone(item.severity)}`}>△</div>
          <div>
            <strong>{item.title}</strong>
            <small>{when(item.timestamp)}</small>
          </div>
          <Status value={item.severity} />
          <Status value={item.status} />
        </div>
      ))}
      {filtered.length === 0 ? (
        <Empty
          title="No incidents in this view"
          copy="Alertmanager events will appear here through the webhook handler."
        />
      ) : null}
    </section>
  );
}
function LogsPage({ pods }: { pods: Pod[] }) {
  const [selected, setSelected] = useState("");
  const [lines, setLines] = useState<string[]>([]);
  const [query, setQuery] = useState("");
  const [busy, setBusy] = useState(false);
  const [controller, setController] = useState<AbortController | null>(null);
  const chosen = pods.find((p) => `${p.namespace}/${p.name}` === selected);
  useEffect(() => () => controller?.abort(), [controller]);
  const load = async () => {
    if (!chosen) return;
    setBusy(true);
    try {
      const data = await api<{ lines: string[] }>(
        `logs?namespace=${encodeURIComponent(
          chosen.namespace,
        )}&pod=${encodeURIComponent(chosen.name)}&lines=200`,
      );
      setLines(data.lines);
    } finally {
      setBusy(false);
    }
  };
  const stream = async () => {
    if (!chosen) return;
    if (controller) {
      controller.abort();
      setController(null);
      return;
    }
    const next = new AbortController();
    setController(next);
    setLines([]);
    const token = sessionStorage.getItem("cloudinferops-token");
    try {
      const response = await fetch(
        `api/logs/stream?namespace=${encodeURIComponent(
          chosen.namespace,
        )}&pod=${encodeURIComponent(chosen.name)}`,
        {
          signal: next.signal,
          headers: token ? { Authorization: `Bearer ${token}` } : {},
        },
      );
      const reader = response.body?.getReader();
      const decoder = new TextDecoder();
      while (reader) {
        const { done, value } = await reader.read();
        if (done) break;
        const chunk = decoder.decode(value, { stream: true });
        setLines((previous) =>
          [...previous, ...chunk.split("\n").filter(Boolean)].slice(-500),
        );
      }
    } catch (e) {
      if (!next.signal.aborted)
        setLines((previous) => [...previous, `Stream ended: ${String(e)}`]);
    } finally {
      setController(null);
    }
  };
  const shown = lines.filter((line) =>
    line.toLowerCase().includes(query.toLowerCase()),
  );
  return (
    <section className="logs-layout">
      <aside className="panel pod-list">
        <h3>Pods</h3>
        {pods.map((p) => (
          <button
            className={selected === `${p.namespace}/${p.name}` ? "active" : ""}
            key={`${p.namespace}/${p.name}`}
            onClick={() => setSelected(`${p.namespace}/${p.name}`)}
          >
            <i className={tone(p.status)} />
            <span>
              <strong>{p.name}</strong>
              <small>{p.namespace}</small>
            </span>
          </button>
        ))}
        {pods.length === 0 ? (
          <Empty title="No pods" copy="Cluster pod inventory is empty." />
        ) : null}
      </aside>
      <div className="panel console">
        <div className="toolbar">
          <input
            aria-label="Filter logs"
            placeholder="Filter log lines…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <div className="button-group">
            <button
              className="secondary"
              disabled={!chosen || busy}
              onClick={() => void load()}
            >
              {busy ? "Loading…" : "Load latest"}
            </button>
            <button disabled={!chosen} onClick={() => void stream()}>
              {controller ? "Stop stream" : "Stream live"}
            </button>
          </div>
        </div>
        <pre>
          {shown.length
            ? shown.map((line, i) => (
                <span key={i}>
                  <b>{String(i + 1).padStart(3, "0")}</b>
                  {line}
                </span>
              ))
            : "Select a pod and load or stream its logs."}
        </pre>
      </div>
    </section>
  );
}
function BenchmarkPage({
  result,
  actions,
  model,
  onRun,
}: {
  result: Benchmark | null;
  actions: boolean;
  model: string;
  onRun: (m: string, n: number) => void;
}) {
  const [target, setTarget] = useState(model);
  const [requests, setRequests] = useState(5);
  return (
    <>
      <section className="panel benchmark-form">
        <div>
          <span className="kicker">INFERENCE PERFORMANCE</span>
          <h3>Run a lightweight readiness benchmark</h3>
          <p>
            Send controlled chat requests through the in-cluster gateway and
            measure end-to-end latency.
          </p>
        </div>
        <div className="action-form">
          <label>
            Model
            <input value={target} onChange={(e) => setTarget(e.target.value)} />
          </label>
          <label>
            Requests
            <input
              type="number"
              min="1"
              max="20"
              value={requests}
              onChange={(e) => setRequests(Number(e.target.value))}
            />
          </label>
          <button disabled={!actions} onClick={() => onRun(target, requests)}>
            Run benchmark
          </button>
        </div>
      </section>
      {result ? (
        <section className="metric-grid benchmark-results">
          <article>
            <span>Average latency</span>
            <strong>{result.average_latency_ms.toFixed(0)} ms</strong>
          </article>
          <article>
            <span>Successful</span>
            <strong>
              {result.succeeded}/{result.requests}
            </strong>
          </article>
          <article>
            <span>Failed</span>
            <strong>{result.failed}</strong>
          </article>
          <article>
            <span>Completed</span>
            <strong className="small-value">{when(result.completed_at)}</strong>
          </article>
        </section>
      ) : (
        <Empty
          title="No benchmark result yet"
          copy="Enable actions and run a benchmark to establish a baseline."
        />
      )}
    </>
  );
}
function ConfigPage({
  value,
  actions,
  onSave,
}: {
  value: PortalConfig;
  actions: boolean;
  onSave: (v: PortalConfig) => void;
}) {
  const [form, setForm] = useState(value);
  useEffect(() => setForm(value), [value]);
  return (
    <section className="panel config">
      <div>
        <span className="kicker">SAFE CONFIGURATION</span>
        <h3>Inference defaults</h3>
        <p>
          Updates the model-config ConfigMap. Secrets are deliberately excluded
          from this portal.
        </p>
      </div>
      <div className="config-grid">
        <label>
          Operational namespace
          <input
            value={form.namespace}
            onChange={(e) => setForm({ ...form, namespace: e.target.value })}
          />
        </label>
        <label>
          Provider
          <select
            value={form.provider}
            onChange={(e) => setForm({ ...form, provider: e.target.value })}
          >
            <option value="ollama">Ollama</option>
            <option value="vllm">vLLM</option>
          </select>
        </label>
        <label>
          Default model
          <input
            value={form.model}
            onChange={(e) => setForm({ ...form, model: e.target.value })}
          />
        </label>
      </div>
      <button disabled={!actions} onClick={() => onSave(form)}>
        Save configuration
      </button>
      <p className="guard">
        {actions
          ? "Changes are applied directly to the in-cluster ConfigMap."
          : "Configuration is read-only while actions are disabled."}
      </p>
    </section>
  );
}
