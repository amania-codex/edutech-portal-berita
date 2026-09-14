/**
 * USER MODEL
 * Handles all database operations for Users.
 * Architecture: MVC - Model layer.
 */
import { getDb } from '../lib/db';

export interface User {
  id: number;
  name: string;
  email: string;
  password: string;
  role: 'admin' | 'user';
  avatar: string | null;
  bio: string | null;
  created_at: string;
  updated_at: string;
}

export interface PublicUser {
  id: number;
  name: string;
  email: string;
  role: 'admin' | 'user';
  avatar: string | null;
  bio: string | null;
  created_at: string;
}

export class UserModel {
  static findByEmail(email: string): User | undefined {
    const db = getDb();
    return db.prepare('SELECT * FROM users WHERE email = ?').get(email) as User | undefined;
  }

  static findById(id: number): PublicUser | undefined {
    const db = getDb();
    return db.prepare(
      'SELECT id, name, email, role, avatar, bio, created_at FROM users WHERE id = ?'
    ).get(id) as PublicUser | undefined;
  }

  static create(data: { name: string; email: string; password: string; role?: string }): number {
    const db = getDb();
    const result = db.prepare(`
      INSERT INTO users (name, email, password, role)
      VALUES (?, ?, ?, ?)
    `).run(data.name, data.email, data.password, data.role || 'user');
    return result.lastInsertRowid as number;
  }

  static update(id: number, data: { name?: string; bio?: string; avatar?: string; password?: string }): void {
    const db = getDb();
    const fields: string[] = [];
    const values: any[] = [];

    if (data.name) { fields.push('name = ?'); values.push(data.name); }
    if (data.bio !== undefined) { fields.push('bio = ?'); values.push(data.bio); }
    if (data.avatar !== undefined) { fields.push('avatar = ?'); values.push(data.avatar); }
    if (data.password) { fields.push('password = ?'); values.push(data.password); }

    if (fields.length === 0) return;
    fields.push('updated_at = CURRENT_TIMESTAMP');
    values.push(id);

    db.prepare(`UPDATE users SET ${fields.join(', ')} WHERE id = ?`).run(...values);
  }

  static getAll(): PublicUser[] {
    const db = getDb();
    return db.prepare(
      'SELECT id, name, email, role, avatar, bio, created_at FROM users ORDER BY created_at DESC'
    ).all() as PublicUser[];
  }

  static count(): number {
    const db = getDb();
    const result = db.prepare('SELECT COUNT(*) as count FROM users').get() as { count: number };
    return result.count;
  }
}
