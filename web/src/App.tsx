import { useEffect, useState } from 'react';

type Overview = {
  cluster: string;
  gateway: string;
  observability: string;
  models: number;
  alerts: number;
  pods: number;
  last_updated: string;
  version: string;
};

type DeploymentStatus = {
  name: string;
  namespace: string;
  status: string;
  replicas: string;
};

type ModelStatus = {
  name: string;
  provider: string;
  status: string;
  location: string;
};

type AlertItem = {
  title: string;
  severity: string;
  status: string;
  timestamp: string;
};

const App = () => {
  const [overview, setOverview] = useState<Overview | null>(null);
  const [deployments, setDeployments] = useState<DeploymentStatus[]>([]);
  const [models, setModels] = useState<ModelStatus[]>([]);
  const [alerts, setAlerts] = useState<AlertItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionMessage, setActionMessage] = useState('');
  const [deployTarget, setDeployTarget] = useState('inference');

  const loadData = async () => {
    setLoading(true);
    const [overviewRes, deploymentsRes, modelsRes, alertsRes] = await Promise.all([
      fetch('/api/overview'),
      fetch('/api/deployments'),
      fetch('/api/models'),
      fetch('/api/alerts'),
    ]);

    setOverview(await overviewRes.json());
    setDeployments(await deploymentsRes.json());
    setModels(await modelsRes.json());
    setAlerts(await alertsRes.json());
    setLoading(false);
  };

  useEffect(() => {
    void loadData();
  }, []);

  const triggerDeploy = async () => {
    const res = await fetch('/api/actions/deploy', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: deployTarget }),
    });
    const payload = await res.json();
    setActionMessage(payload.message ?? 'Deployment request submitted');
    await loadData();
  };

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div>
          <h1>CloudInferOps</h1>
          <p>Self-hosted AI inference operations</p>
        </div>
        <nav>
          <a href="#overview">Overview</a>
          <a href="#deployments">Deployments</a>
          <a href="#models">Models</a>
          <a href="#alerts">Alerts</a>
        </nav>
      </aside>

      <main className="main-content">
        <section id="overview" className="hero-card">
          <div>
            <p className="eyebrow">Operational dashboard</p>
            <h2>Run inference workloads with clarity</h2>
            <p>Deploy, monitor, and observe your LLM stack from one self-hosted portal.</p>
          </div>
          <div className="hero-stats">
            <div className="stat-card">
              <span>Cluster</span>
              <strong>{overview?.cluster ?? '—'}</strong>
            </div>
            <div className="stat-card">
              <span>Gateway</span>
              <strong>{overview?.gateway ?? '—'}</strong>
            </div>
            <div className="stat-card">
              <span>Models</span>
              <strong>{overview?.models ?? 0}</strong>
            </div>
          </div>
        </section>

        <section className="panel action-panel">
          <div className="action-header">
            <h3>Control plane actions</h3>
            <p>Trigger a deploy request and refresh the dashboard instantly.</p>
          </div>
          <div className="action-controls">
            <input value={deployTarget} onChange={(event) => setDeployTarget(event.target.value)} placeholder="deployment target" />
            <button onClick={() => void triggerDeploy()}>Deploy</button>
          </div>
          {actionMessage ? <div className="action-message">{actionMessage}</div> : null}
        </section>

        <section className="grid">
          <div className="panel">
            <h3>System health</h3>
            {loading ? <div className="empty-state">Loading live status…</div> : (
              <ul>
                <li>Observability: {overview?.observability ?? '—'}</li>
                <li>Pods: {overview?.pods ?? 0}</li>
                <li>Last updated: {overview?.last_updated ?? '—'}</li>
              </ul>
            )}
          </div>

          <div className="panel" id="deployments">
            <h3>Deployments</h3>
            {loading ? <div className="empty-state">Loading deployments…</div> : deployments.length === 0 ? <div className="empty-state">No deployments detected yet.</div> : (
              <table>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Namespace</th>
                    <th>Status</th>
                    <th>Replicas</th>
                  </tr>
                </thead>
                <tbody>
                  {deployments.map((deployment) => (
                    <tr key={deployment.name}>
                      <td>{deployment.name}</td>
                      <td>{deployment.namespace}</td>
                      <td>{deployment.status}</td>
                      <td>{deployment.replicas}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </section>

        <section className="grid">
          <div className="panel" id="models">
            <h3>Models</h3>
            {models.length === 0 ? <div className="empty-state">No model inventory found.</div> : (
              <table>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Provider</th>
                    <th>Status</th>
                    <th>Location</th>
                  </tr>
                </thead>
                <tbody>
                  {models.map((model) => (
                    <tr key={model.name}>
                      <td>{model.name}</td>
                      <td>{model.provider}</td>
                      <td>{model.status}</td>
                      <td>{model.location}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          <div className="panel" id="alerts">
            <h3>Recent alerts</h3>
            {alerts.length === 0 ? <div className="empty-state">No alerts detected.</div> : (
              <ul>
                {alerts.map((alert) => (
                  <li key={alert.title}>
                    <strong>{alert.title}</strong>
                    <div>{alert.severity} • {alert.status}</div>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </section>
      </main>
    </div>
  );
};

export default App;
