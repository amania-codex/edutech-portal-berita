// API Route: GET /api/news?limit=&offset=&status=&category=
//            POST /api/news (create - admin only)
import type { APIRoute } from 'astro';
import { NewsController } from '../../../controllers/NewsController';
import { getSessionFromRequest } from '../../../lib/session';

export const GET: APIRoute = async ({ request }) => {
  const url = new URL(request.url);
  const limit = parseInt(url.searchParams.get('limit') || '12');
  const offset = parseInt(url.searchParams.get('offset') || '0');
  const status = url.searchParams.get('status') || undefined;
  const category = url.searchParams.get('category') || undefined;
  const query = url.searchParams.get('q') || undefined;

  if (query) {
    const result = NewsController.search(query);
    return new Response(JSON.stringify(result.data || []), {
      headers: { 'Content-Type': 'application/json' },
    });
  }

  const result = NewsController.getAll({ limit, offset, status, category });
  return new Response(JSON.stringify(result.data || []), {
    headers: { 'Content-Type': 'application/json' },
  });
};

export const POST: APIRoute = async ({ request }) => {
  const session = await getSessionFromRequest(request);
  if (!session || session.role !== 'admin') {
    return new Response(JSON.stringify({ error: 'Unauthorized.' }), {
      status: 401, headers: { 'Content-Type': 'application/json' },
    });
  }

  try {
    const body = await request.json();
    const result = NewsController.create({
      ...body,
      author_id: session.userId,
      author: body.author || session.name,
    });

    if (!result.success) {
      return new Response(JSON.stringify({ error: result.error }), {
        status: 400, headers: { 'Content-Type': 'application/json' },
      });
    }

    return new Response(JSON.stringify({ id: result.id }), {
      status: 201, headers: { 'Content-Type': 'application/json' },
    });
  } catch {
    return new Response(JSON.stringify({ error: 'Server error.' }), {
      status: 500, headers: { 'Content-Type': 'application/json' },
    });
  }
};
