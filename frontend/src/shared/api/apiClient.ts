const BASE_URL = (import.meta.env.VITE_API_URL as string | undefined) || 'http://localhost:8080';

export const apiClient = {
  get: async (endpoint: string) => {
    const response = await fetch(`${BASE_URL}${endpoint}`);
    if (!response.ok) throw new Error('Network response was not ok');
    return response.json();
  },
  post: async (endpoint: string, data: any) => {
    const response = await fetch(`${BASE_URL}${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    if (!response.ok) throw new Error('Network response was not ok');
    return response.json();
  },
  delete: async (endpoint: string) => {
    const response = await fetch(`${BASE_URL}${endpoint}`, {
      method: 'DELETE',
    });
    if (!response.ok) throw new Error('Network response was not ok');
    return response.json();
  },
  rawFetch: async (endpoint: string, options: RequestInit = {}) => {
    return fetch(`${BASE_URL}${endpoint}`, {
      ...options,
      headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    });
  },
  BASE_URL,
};

export { BASE_URL };