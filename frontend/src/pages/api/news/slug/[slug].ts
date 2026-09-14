// API Route: GET /api/news/slug/[slug]
import type { APIRoute } from 'astro';
import { NewsController } from '../../../../controllers/NewsController';

export const GET: APIRoute = async ({ params }) => {
  const { slug } = params;
  const result = NewsController.getBySlug(slug || '');

  if (!result.success) {
    return new Response(JSON.stringify({ error: result.error }), {
      status: 404, headers: { 'Content-Type': 'application/json' },
    });
  }

  return new Response(JSON.stringify(result.data), {
    headers: { 'Content-Type': 'application/json' },
  });
};
