// API Route: GET /api/news/trending
import type { APIRoute } from 'astro';
import { NewsController } from '../../../controllers/NewsController';

export const GET: APIRoute = async ({ request }) => {
  const url = new URL(request.url);
  const limit = parseInt(url.searchParams.get('limit') || '5');
  const result = NewsController.getTrending(limit);
  return new Response(JSON.stringify(result.data || []), {
    headers: { 'Content-Type': 'application/json' },
  });
};
