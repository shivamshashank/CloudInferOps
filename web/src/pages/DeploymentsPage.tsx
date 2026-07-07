import { useState } from "react";
import { Deployment } from "../types";
import Status from "../components/Status";
import Empty from "../components/Empty";
import { when } from "../utils";

export default function DeploymentsPage({
  items,
  actions,
  onDeploy,
  onUndeploy,
  onRestart,
}: {
  items: Deployment[];
  actions: boolean;
  onDeploy: (x: string) => void;
  onUndeploy: (x: string) => void;
  onRestart: (n: string, x: string) => void;
}) {
  const [query, setQuery] = useState("");
  const filtered = items.filter((i) =>
    `${i.name} ${i.namespace}`.toLowerCase().includes(query.toLowerCase()),
  );
  const isInferenceDeployed = items.some((i) => i.namespace === "inference");
  const isObservabilityDeployed = items.some(
    (i) => i.namespace === "observability" && i.name !== "cloudinferops-ui",
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
          {isInferenceDeployed ? (
            <button
              className="secondary"
              disabled={!actions}
              onClick={() => onUndeploy("inference")}
            >
              Undeploy inference
            </button>
          ) : (
            <button disabled={!actions} onClick={() => onDeploy("inference")}>
              Deploy inference
            </button>
          )}
          {isObservabilityDeployed ? (
            <button
              className="secondary"
              disabled={!actions}
              onClick={() => onUndeploy("observability")}
            >
              Undeploy observability
            </button>
          ) : (
            <button
              disabled={!actions}
              onClick={() => onDeploy("observability")}
            >
              Deploy observability
            </button>
          )}
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
