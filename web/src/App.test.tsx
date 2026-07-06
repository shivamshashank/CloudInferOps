import { render, screen, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import App from './App';

describe('App Component', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn().mockImplementation((url: string) => {
      if (url.includes('/api/overview')) {
        return Promise.resolve({
          json: () => Promise.resolve({
            cluster: 'test-cluster',
            gateway: 'test-gateway',
            observability: 'active',
            models: 2,
            alerts: 0,
            pods: 4,
            last_updated: '2026-07-06T16:00:00Z',
            version: 'v0.1.0'
          })
        });
      }
      if (url.includes('/api/deployments')) {
        return Promise.resolve({
          json: () => Promise.resolve([
            { name: 'llama-3', namespace: 'default', status: 'Running', replicas: '1/1' }
          ])
        });
      }
      if (url.includes('/api/models')) {
        return Promise.resolve({
          json: () => Promise.resolve([
            { name: 'llama-3', provider: 'ollama', status: 'loaded', location: 'local' }
          ])
        });
      }
      if (url.includes('/api/alerts')) {
        return Promise.resolve({
          json: () => Promise.resolve([])
        });
      }
      return Promise.reject(new Error('Unknown URL: ' + url));
    }));
  });

  it('renders sidebar and dashboard information', async () => {
    await act(async () => {
      render(<App />);
    });

    // Check for the title
    expect(screen.getByText('CloudInferOps')).toBeInTheDocument();
    expect(screen.getByText('Self-hosted AI inference operations')).toBeInTheDocument();

    // Check if the mock data shows up
    expect(await screen.findByText('test-cluster')).toBeInTheDocument();
    expect(await screen.findByText('test-gateway')).toBeInTheDocument();
  });
});
