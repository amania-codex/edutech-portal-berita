/**
 * AUTH CONTROLLER
 * Handles authentication business logic (login, register, logout).
 * Architecture: MVC - Controller layer.
 */
import bcrypt from 'bcryptjs';
import { UserModel } from '../models/UserModel';
import { createSession, verifySession } from '../lib/session';

export interface AuthResult {
  success: boolean;
  error?: string;
  token?: string;
  user?: { id: number; name: string; email: string; role: string };
}

export class AuthController {
  static async login(email: string, password: string): Promise<AuthResult> {
    if (!email || !password) {
      return { success: false, error: 'Email dan password wajib diisi.' };
    }

    const user = UserModel.findByEmail(email.toLowerCase().trim());
    if (!user) {
      return { success: false, error: 'Email atau password salah.' };
    }

    const isValid = bcrypt.compareSync(password, user.password);
    if (!isValid) {
      return { success: false, error: 'Email atau password salah.' };
    }

    const token = await createSession({
      userId: user.id,
      email: user.email,
      name: user.name,
      role: user.role as 'admin' | 'user',
    });

    return {
      success: true,
      token,
      user: { id: user.id, name: user.name, email: user.email, role: user.role },
    };
  }

  static async register(name: string, email: string, password: string): Promise<AuthResult> {
    if (!name || !email || !password) {
      return { success: false, error: 'Semua field wajib diisi.' };
    }
    if (name.trim().length < 2) {
      return { success: false, error: 'Nama minimal 2 karakter.' };
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      return { success: false, error: 'Format email tidak valid.' };
    }
    if (password.length < 6) {
      return { success: false, error: 'Password minimal 6 karakter.' };
    }

    const existingUser = UserModel.findByEmail(email.toLowerCase().trim());
    if (existingUser) {
      return { success: false, error: 'Email sudah terdaftar.' };
    }

    const hashedPassword = bcrypt.hashSync(password, 10);
    const userId = UserModel.create({
      name: name.trim(),
      email: email.toLowerCase().trim(),
      password: hashedPassword,
      role: 'user',
    });

    const token = await createSession({
      userId,
      email: email.toLowerCase().trim(),
      name: name.trim(),
      role: 'user',
    });

    return {
      success: true,
      token,
      user: { id: userId, name: name.trim(), email: email.toLowerCase().trim(), role: 'user' },
    };
  }
}
