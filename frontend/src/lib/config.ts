/**
 * Global App Configuration
 * Allows easy switching between development (localhost:3000) and production API host.
 */
export const BACKEND_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:3000';
export const API_BASE = `${BACKEND_URL}/api`;
