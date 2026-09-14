/**
 * NEWS MODEL
 * Handles all database operations for News articles.
 * Architecture: MVC - Model layer.
 */
import { getDb } from '../lib/db';

export interface NewsArticle {
  id: number;
  title: string;
  slug: string;
  excerpt: string | null;
  content: string;
  image_url: string | null;
  category: string;
  tags: string | null;
  author_id: number;
  author: string;
  status: 'draft' | 'published';
  is_featured: number;
  view_count: number;
  created_at: string;
  updated_at: string;
}

export interface CreateNewsData {
  title: string;
  slug: string;
  excerpt?: string;
  content: string;
  image_url?: string;
  category: string;
  tags?: string;
  author_id: number;
  author: string;
  status?: 'draft' | 'published';
  is_featured?: number;
}

export class NewsModel {
  static findBySlug(slug: string): NewsArticle | undefined {
    const db = getDb();
    return db.prepare('SELECT * FROM news WHERE slug = ? AND status = ?').get(slug, 'published') as NewsArticle | undefined;
  }

  static findById(id: number): NewsArticle | undefined {
    const db = getDb();
    return db.prepare('SELECT * FROM news WHERE id = ?').get(id) as NewsArticle | undefined;
  }

  static getAll(options: { limit?: number; offset?: number; status?: string; category?: string } = {}): NewsArticle[] {
    const db = getDb();
    const { limit = 20, offset = 0, status, category } = options;
    
    let query = 'SELECT * FROM news WHERE 1=1';
    const params: any[] = [];
    
    if (status) { query += ' AND status = ?'; params.push(status); }
    if (category) { query += ' AND category = ?'; params.push(category); }
    
    query += ' ORDER BY created_at DESC LIMIT ? OFFSET ?';
    params.push(limit, offset);
    
    return db.prepare(query).all(...params) as NewsArticle[];
  }

  static getFeatured(limit = 5): NewsArticle[] {
    const db = getDb();
    return db.prepare(
      'SELECT * FROM news WHERE is_featured = 1 AND status = ? ORDER BY created_at DESC LIMIT ?'
    ).all('published', limit) as NewsArticle[];
  }

  static getTrending(limit = 5): NewsArticle[] {
    const db = getDb();
    return db.prepare(
      'SELECT * FROM news WHERE status = ? ORDER BY view_count DESC, created_at DESC LIMIT ?'
    ).all('published', limit) as NewsArticle[];
  }

  static getByCategory(category: string, limit = 12, offset = 0): NewsArticle[] {
    const db = getDb();
    return db.prepare(
      'SELECT * FROM news WHERE category = ? AND status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?'
    ).all(category, 'published', limit, offset) as NewsArticle[];
  }

  static search(query: string, limit = 12): NewsArticle[] {
    const db = getDb();
    const searchTerm = `%${query}%`;
    return db.prepare(
      'SELECT * FROM news WHERE status = ? AND (title LIKE ? OR excerpt LIKE ? OR content LIKE ?) ORDER BY created_at DESC LIMIT ?'
    ).all('published', searchTerm, searchTerm, searchTerm, limit) as NewsArticle[];
  }

  static create(data: CreateNewsData): number {
    const db = getDb();
    const result = db.prepare(`
      INSERT INTO news (title, slug, excerpt, content, image_url, category, tags, author_id, author, status, is_featured)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `).run(
      data.title, data.slug, data.excerpt || null, data.content,
      data.image_url || null, data.category, data.tags || null,
      data.author_id, data.author, data.status || 'draft', data.is_featured || 0
    );
    return result.lastInsertRowid as number;
  }

  static update(id: number, data: Partial<CreateNewsData> & { status?: string; is_featured?: number }): void {
    const db = getDb();
    const fields: string[] = [];
    const values: any[] = [];

    if (data.title) { fields.push('title = ?'); values.push(data.title); }
    if (data.slug) { fields.push('slug = ?'); values.push(data.slug); }
    if (data.excerpt !== undefined) { fields.push('excerpt = ?'); values.push(data.excerpt); }
    if (data.content) { fields.push('content = ?'); values.push(data.content); }
    if (data.image_url !== undefined) { fields.push('image_url = ?'); values.push(data.image_url); }
    if (data.category) { fields.push('category = ?'); values.push(data.category); }
    if (data.tags !== undefined) { fields.push('tags = ?'); values.push(data.tags); }
    if (data.status) { fields.push('status = ?'); values.push(data.status); }
    if (data.is_featured !== undefined) { fields.push('is_featured = ?'); values.push(data.is_featured); }

    if (fields.length === 0) return;
    fields.push('updated_at = CURRENT_TIMESTAMP');
    values.push(id);

    db.prepare(`UPDATE news SET ${fields.join(', ')} WHERE id = ?`).run(...values);
  }

  static delete(id: number): void {
    const db = getDb();
    db.prepare('DELETE FROM news WHERE id = ?').run(id);
  }

  static incrementViews(id: number): void {
    const db = getDb();
    db.prepare('UPDATE news SET view_count = view_count + 1 WHERE id = ?').run(id);
  }

  static count(status?: string): number {
    const db = getDb();
    if (status) {
      const result = db.prepare('SELECT COUNT(*) as count FROM news WHERE status = ?').get(status) as { count: number };
      return result.count;
    }
    const result = db.prepare('SELECT COUNT(*) as count FROM news').get() as { count: number };
    return result.count;
  }

  static generateSlug(title: string): string {
    return title
      .toLowerCase()
      .replace(/[àáâãäå]/g, 'a').replace(/[èéêë]/g, 'e').replace(/[ìíîï]/g, 'i')
      .replace(/[òóôõö]/g, 'o').replace(/[ùúûü]/g, 'u')
      .replace(/[^a-z0-9\s-]/g, '')
      .replace(/\s+/g, '-')
      .replace(/-+/g, '-')
      .trim();
  }
}
