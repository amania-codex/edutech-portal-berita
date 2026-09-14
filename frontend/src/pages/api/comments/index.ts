// API Route: GET /api/comments?news_id=&flaggedOnly=
//            POST /api/comments (create - authenticated users)
//            DELETE /api/comments?id= (admin only)
//            PATCH /api/comments?id=&action=unflag (admin only)
import type { APIRoute } from 'astro';
import { CommentController } from '../../../controllers/CommentController';
import { getSessionFromRequest } from '../../../lib/session';

export const GET: APIRoute = async ({ request }) => {
  const url = new URL(request.url);
  const newsId = parseInt(url.searchParams.get('news_id') || '0');
  const flaggedOnly = url.searchParams.get('flaggedOnly') === 'true';

  if (newsId) {
    const result = CommentController.getByNewsId(newsId);
    return new Response(JSON.stringify(result.data || []), {
      headers: { 'Content-Type': 'application/json' },
    });
  }

  // Admin: get all comments with stats
  const result = CommentController.getAll({ flaggedOnly });
  return new Response(JSON.stringify(result.data || {}), {
    headers: { 'Content-Type': 'application/json' },
  });
};

export const POST: APIRoute = async ({ request }) => {
  const session = await getSessionFromRequest(request);
  if (!session) {
    return new Response(JSON.stringify({ error: 'Login terlebih dahulu untuk berkomentar.' }), {
      status: 401, headers: { 'Content-Type': 'application/json' },
    });
  }

  try {
    const body = await request.json();
    const result = CommentController.create({
      news_id: body.news_id,
      user_id: session.userId,
      user_name: session.name,
      content: body.content,
    });

    if (!result.success) {
      return new Response(JSON.stringify({ error: result.error }), {
        status: 400, headers: { 'Content-Type': 'application/json' },
      });
    }

    return new Response(JSON.stringify(result.data), {
      status: 201, headers: { 'Content-Type': 'application/json' },
    });
  } catch {
    return new Response(JSON.stringify({ error: 'Server error.' }), {
      status: 500, headers: { 'Content-Type': 'application/json' },
    });
  }
};

export const DELETE: APIRoute = async ({ request }) => {
  const session = await getSessionFromRequest(request);
  if (!session || session.role !== 'admin') {
    return new Response(JSON.stringify({ error: 'Unauthorized.' }), {
      status: 401, headers: { 'Content-Type': 'application/json' },
    });
  }

  const url = new URL(request.url);
  const id = parseInt(url.searchParams.get('id') || '0');
  const result = CommentController.delete(id);

  return new Response(JSON.stringify(result.success ? { success: true } : { error: result.error }), {
    status: result.success ? 200 : 404,
    headers: { 'Content-Type': 'application/json' },
  });
};

export const PATCH: APIRoute = async ({ request }) => {
  const session = await getSessionFromRequest(request);
  if (!session || session.role !== 'admin') {
    return new Response(JSON.stringify({ error: 'Unauthorized.' }), {
      status: 401, headers: { 'Content-Type': 'application/json' },
    });
  }

  const url = new URL(request.url);
  const id = parseInt(url.searchParams.get('id') || '0');
  const action = url.searchParams.get('action');

  if (action === 'unflag') {
    const result = CommentController.unflag(id);
    return new Response(JSON.stringify(result.success ? { success: true } : { error: result.error }), {
      headers: { 'Content-Type': 'application/json' },
    });
  }

  return new Response(JSON.stringify({ error: 'Invalid action.' }), {
    status: 400, headers: { 'Content-Type': 'application/json' },
  });
};
