// API Route: GET /api/news/featured
import type { APIRoute } from 'astro';
import { NewsController } from '../../../controllers/NewsController';

export const GET: APIRoute = async ({ request }) => {
  const url = new URL(request.url);
  const limit = parseInt(url.searchParams.get('limit') || '5');
  const result = NewsController.getFeatured(limit);
  return new Response(JSON.stringify(result.data || []), {
    headers: { 'Content-Type': 'application/json' },
  });
};
