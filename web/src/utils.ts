export async function api<T>(path: string, options?: RequestInit): Promise<T> {
  const token = sessionStorage.getItem("cloudinferops-token");
  // Relative API call so that it resolves under /cloudinferops/api/ correctly
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

export const tone = (status: string) =>
  /running|healthy|connected|active|available|ok/i.test(status)
    ? "good"
    : /partial|pending|warning|unknown|checking/i.test(status)
      ? "warn"
      : "muted";

export const when = (value: string) =>
  value ? new Date(value).toLocaleString() : "—";
