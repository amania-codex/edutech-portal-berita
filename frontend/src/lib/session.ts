/**
 * SESSION MANAGEMENT
 * JWT-based session handling for authentication.
 * Architecture: MVC - Auth utility/helper layer.
 */
import { SignJWT, jwtVerify } from 'jose';

const SECRET_KEY = new TextEncoder().encode(
  process.env.JWT_SECRET || 'edutech-super-secret-key-change-in-production-2026'
);

export interface SessionPayload {
  userId: number;
  email: string;
  name: string;
  role: 'admin' | 'user';
}

export async function createSession(payload: SessionPayload): Promise<string> {
  return new SignJWT({ ...payload })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt()
    .setExpirationTime('7d')
    .sign(SECRET_KEY);
}

export async function verifySession(token: string): Promise<SessionPayload | null> {
  try {
    const { payload } = await jwtVerify(token, SECRET_KEY);
    return payload as unknown as SessionPayload;
  } catch {
    return null;
  }
}

export async function getSessionFromRequest(request: Request): Promise<SessionPayload | null> {
  const cookie = request.headers.get('cookie') || '';
  const tokenMatch = cookie.match(/auth_token=([^;]+)/);
  if (!tokenMatch) return null;
  return verifySession(tokenMatch[1]);
}

export function createAuthCookie(token: string): string {
  return `auth_token=${token}; Path=/; HttpOnly; SameSite=Lax; Max-Age=${7 * 24 * 60 * 60}`;
}

export function clearAuthCookie(): string {
  return `auth_token=; Path=/; HttpOnly; SameSite=Lax; Max-Age=0`;
}
