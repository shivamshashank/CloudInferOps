import { act, fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";

const fixtures: Record<string, unknown> = {
  overview: {
    cluster: "Connected",
    namespace: "observability",
    gateway: "Running",
    observability: "Healthy",
    models: 1,
    alerts: 0,
    pods: 4,
    last_updated: "2026-07-06T16:00:00Z",
    version: "v1.0.8",
    actions_enabled: false,
  },
  deployments: [
    {
      name: "gateway-deployment",
      namespace: "inference",
      status: "Running",
      replicas: "1/1",
      last_updated: "2026-07-06T16:00:00Z",
    },
  ],
  models: [
    {
      name: "llama3",
      provider: "ollama",
      status: "Available",
      location: "Cluster",
    },
  ],
  alerts: [],
  observability: [
    { name: "grafana", namespace: "observability", status: "Running" },
  ],
  pods: [{ name: "gateway-1", namespace: "inference", status: "Running" }],
  config: { namespace: "observability", provider: "ollama", model: "llama3" },
  benchmark: null,
};

describe("CloudInferOps portal", () => {
  beforeEach(() => {
    location.hash = "";
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation((url: string) => {
        const key = url.replace(/^api\//, "").split("?")[0];
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(fixtures[key]),
        });
      }),
    );
  });
  it("loads live overview and navigates across portal pages", async () => {
    await act(async () => render(<App />));
    expect((await screen.findAllByText("Connected")).length).toBeGreaterThan(0);
    expect(screen.getByText("gateway-deployment")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /Models/i }));
    expect(await screen.findByText("llama3")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /Logs/i }));
    expect(await screen.findByText("gateway-1")).toBeInTheDocument();
  });
  it("clearly communicates read-only action mode", async () => {
    await act(async () => render(<App />));
    fireEvent.click(screen.getByRole("button", { name: /Deployments/i }));
    expect(await screen.findByText(/Read-only mode/)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Undeploy inference" }),
    ).toBeDisabled();
  });
});
