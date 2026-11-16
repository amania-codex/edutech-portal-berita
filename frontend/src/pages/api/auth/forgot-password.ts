// API Route: POST /api/auth/forgot-password
// Proxy ke Go Fiber backend
import type { APIRoute } from 'astro';
import { API_BASE } from '../../../lib/config';

export const POST: APIRoute = async ({ request }) => {
  try {
    const body = await request.json();
    const res = await fetch(`${API_BASE}/auth/forgot-password`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    const data = await res.json();
    return new Response(JSON.stringify(data), {
      status: res.status,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch {
    return new Response(JSON.stringify({ success: false, error: 'Server error.' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
};
