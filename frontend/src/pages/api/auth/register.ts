// API Route: POST /api/auth/register
// Proxy ke Go Fiber backend
import type { APIRoute } from 'astro';

const GO_API = 'http://localhost:3000/api';

export const POST: APIRoute = async ({ request }) => {
  try {
    const body = await request.json();
    const res = await fetch(`${GO_API}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    const data = await res.json();
    const headers = new Headers({ 'Content-Type': 'application/json' });
    if (data.token) {
      headers.set('Set-Cookie', `auth_token=${data.token}; Path=/; HttpOnly; SameSite=Lax; Max-Age=${7 * 24 * 3600}`);
    }
    return new Response(JSON.stringify(data), { status: res.status, headers });
  } catch {
    return new Response(JSON.stringify({ error: 'Server error.' }), { status: 500, headers: { 'Content-Type': 'application/json' } });
  }
};