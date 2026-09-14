/**
 * COMMENT MODEL
 * Handles all database operations for Comments.
 * Architecture: MVC - Model layer.
 */
import { getDb } from '../lib/db';

export interface Comment {
  id: number;
  news_id: number;
  user_id: number;
  user_name: string;
  content: string;
  sentiment: 'positive' | 'negative' | 'neutral';
  sentiment_score: number;
  is_flagged: number;
  created_at: string;
}

export interface CreateCommentData {
  news_id: number;
  user_id: number;
  user_name: string;
  content: string;
  sentiment?: string;
  sentiment_score?: number;
  is_flagged?: number;
}

export class CommentModel {
  static findByNewsId(newsId: number): Comment[] {
    const db = getDb();
    return db.prepare(
      'SELECT * FROM comments WHERE news_id = ? ORDER BY created_at DESC'
    ).all(newsId) as Comment[];
  }

  static findById(id: number): Comment | undefined {
    const db = getDb();
    return db.prepare('SELECT * FROM comments WHERE id = ?').get(id) as Comment | undefined;
  }

  static create(data: CreateCommentData): number {
    const db = getDb();
    const result = db.prepare(`
      INSERT INTO comments (news_id, user_id, user_name, content, sentiment, sentiment_score, is_flagged)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `).run(
      data.news_id, data.user_id, data.user_name, data.content,
      data.sentiment || 'neutral', data.sentiment_score || 0, data.is_flagged || 0
    );
    return result.lastInsertRowid as number;
  }

  static delete(id: number): void {
    const db = getDb();
    db.prepare('DELETE FROM comments WHERE id = ?').run(id);
  }

  static getAll(options: { limit?: number; flaggedOnly?: boolean } = {}): Comment[] {
    const db = getDb();
    const { limit = 50, flaggedOnly = false } = options;
    
    if (flaggedOnly) {
      return db.prepare(
        'SELECT c.*, n.title as news_title FROM comments c LEFT JOIN news n ON c.news_id = n.id WHERE c.is_flagged = 1 ORDER BY c.created_at DESC LIMIT ?'
      ).all(limit) as Comment[];
    }
    
    return db.prepare(
      'SELECT c.*, n.title as news_title FROM comments c LEFT JOIN news n ON c.news_id = n.id ORDER BY c.created_at DESC LIMIT ?'
    ).all(limit) as Comment[];
  }

  static count(): number {
    const db = getDb();
    const result = db.prepare('SELECT COUNT(*) as count FROM comments').get() as { count: number };
    return result.count;
  }

  static countFlagged(): number {
    const db = getDb();
    const result = db.prepare('SELECT COUNT(*) as count FROM comments WHERE is_flagged = 1').get() as { count: number };
    return result.count;
  }

  static getSentimentStats(): { positive: number; negative: number; neutral: number } {
    const db = getDb();
    const rows = db.prepare('SELECT sentiment, COUNT(*) as count FROM comments GROUP BY sentiment').all() as { sentiment: string; count: number }[];
    const stats = { positive: 0, negative: 0, neutral: 0 };
    for (const row of rows) {
      if (row.sentiment === 'positive') stats.positive = row.count;
      else if (row.sentiment === 'negative') stats.negative = row.count;
      else stats.neutral = row.count;
    }
    return stats;
  }

  static unflag(id: number): void {
    const db = getDb();
    db.prepare('UPDATE comments SET is_flagged = 0 WHERE id = ?').run(id);
  }
}
