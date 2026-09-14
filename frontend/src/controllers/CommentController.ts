/**
 * COMMENT CONTROLLER
 * Handles comment creation and analysis business logic.
 * Architecture: MVC - Controller layer.
 */
import { CommentModel, type CreateCommentData } from '../models/CommentModel';
import { analyzeSentiment } from '../lib/sentimentAnalysis';

export interface CommentResult {
  success: boolean;
  error?: string;
  data?: any;
  id?: number;
}

export class CommentController {
  static getByNewsId(newsId: number): CommentResult {
    try {
      const comments = CommentModel.findByNewsId(newsId);
      return { success: true, data: comments };
    } catch {
      return { success: false, error: 'Gagal mengambil komentar.' };
    }
  }

  static create(data: { news_id: number; user_id: number; user_name: string; content: string }): CommentResult {
    if (!data.content?.trim()) return { success: false, error: 'Komentar tidak boleh kosong.' };
    if (data.content.trim().length < 3) return { success: false, error: 'Komentar terlalu pendek.' };
    if (data.content.length > 1000) return { success: false, error: 'Komentar terlalu panjang (maks 1000 karakter).' };

    // Run sentiment analysis
    const analysis = analyzeSentiment(data.content);

    try {
      const id = CommentModel.create({
        ...data,
        content: data.content.trim(),
        sentiment: analysis.sentiment,
        sentiment_score: analysis.score,
        is_flagged: analysis.isFlagged ? 1 : 0,
      });
      const comment = CommentModel.findById(id);
      return { success: true, id, data: comment };
    } catch {
      return { success: false, error: 'Gagal menyimpan komentar.' };
    }
  }

  static delete(id: number): CommentResult {
    const existing = CommentModel.findById(id);
    if (!existing) return { success: false, error: 'Komentar tidak ditemukan.' };
    try {
      CommentModel.delete(id);
      return { success: true };
    } catch {
      return { success: false, error: 'Gagal menghapus komentar.' };
    }
  }

  static getAll(options: { flaggedOnly?: boolean } = {}): CommentResult {
    try {
      const comments = CommentModel.getAll({ limit: 100, flaggedOnly: options.flaggedOnly });
      const stats = CommentModel.getSentimentStats();
      return { success: true, data: { comments, stats } };
    } catch {
      return { success: false, error: 'Gagal mengambil data komentar.' };
    }
  }

  static unflag(id: number): CommentResult {
    try {
      CommentModel.unflag(id);
      return { success: true };
    } catch {
      return { success: false, error: 'Gagal mengubah status komentar.' };
    }
  }
}
