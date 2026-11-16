// API Route: POST /api/events/[id]/register
// Proxy to Go Fiber backend
import type { APIRoute } from 'astro';
import { API_BASE } from '../../../../lib/config';

export const POST: APIRoute = async ({ params, request }) => {
  const { id } = params;
  try {
    const cookie = request.headers.get('cookie') || '';
    const body = await request.json();

    const res = await fetch(`${API_BASE}/events/${id}/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Cookie': cookie,
      },
      body: JSON.stringify(body),
    });

    const data = await res.json();
    return new Response(JSON.stringify(data), {
      status: res.status,
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (err: any) {
    return new Response(JSON.stringify({ success: false, error: 'Gagal menghubungi server pendaftaran: ' + (err?.message || 'Error') }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' },
    });
  }
};
