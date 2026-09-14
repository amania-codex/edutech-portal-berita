// API Route: GET /api/news/[id]
//            PUT /api/news/[id] (update - admin only)
//            DELETE /api/news/[id] (delete - admin only)
import type { APIRoute } from 'astro';
import { NewsController } from '../../../controllers/NewsController';
import { getSessionFromRequest } from '../../../lib/session';

export const GET: APIRoute = async ({ params }) => {
  const id = parseInt(params.id || '0');
  const result = NewsController.getById(id);

  if (!result.success) {
    return new Response(JSON.stringify({ error: result.error }), {
      status: 404, headers: { 'Content-Type': 'application/json' },
    });
  }

  return new Response(JSON.stringify(result.data), {
    headers: { 'Content-Type': 'application/json' },
  });
};

export const PUT: APIRoute = async ({ request, params }) => {
  const session = await getSessionFromRequest(request);
  if (!session || session.role !== 'admin') {
    return new Response(JSON.stringify({ error: 'Unauthorized.' }), {
      status: 401, headers: { 'Content-Type': 'application/json' },
    });
  }

  const id = parseInt(params.id || '0');
  try {
    const body = await request.json();
    const result = NewsController.update(id, body);

    if (!result.success) {
      return new Response(JSON.stringify({ error: result.error }), {
        status: 400, headers: { 'Content-Type': 'application/json' },
      });
    }

    return new Response(JSON.stringify(result.data), {
      headers: { 'Content-Type': 'application/json' },
    });
  } catch {
    return new Response(JSON.stringify({ error: 'Server error.' }), {
      status: 500, headers: { 'Content-Type': 'application/json' },
    });
  }
};

export const DELETE: APIRoute = async ({ request, params }) => {
  const session = await getSessionFromRequest(request);
  if (!session || session.role !== 'admin') {
    return new Response(JSON.stringify({ error: 'Unauthorized.' }), {
      status: 401, headers: { 'Content-Type': 'application/json' },
    });
  }

  const id = parseInt(params.id || '0');
  const result = NewsController.delete(id);

  if (!result.success) {
    return new Response(JSON.stringify({ error: result.error }), {
      status: 404, headers: { 'Content-Type': 'application/json' },
    });
  }

  return new Response(JSON.stringify({ success: true }), {
    headers: { 'Content-Type': 'application/json' },
  });
};
