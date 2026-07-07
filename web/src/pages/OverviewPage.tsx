import { Overview, Deployment, Alert, Page } from "../types";
import Status from "../components/Status";
import Empty from "../components/Empty";
import { when } from "../utils";

export default function OverviewPage({
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
