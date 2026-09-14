/**
 * DATABASE LAYER (Model Foundation)
 * Initializes SQLite database with all required tables.
 * Architecture: MVC - this is the database connection singleton.
 */
import Database from 'better-sqlite3';
import { mkdirSync } from 'fs';
import { join } from 'path';
import bcrypt from 'bcryptjs';

const DB_PATH = join(process.cwd(), 'data', 'newsportal.db');

let db: Database.Database | null = null;

export function getDb(): Database.Database {
  if (!db) {
    mkdirSync(join(process.cwd(), 'data'), { recursive: true });
    db = new Database(DB_PATH);
    db.pragma('journal_mode = WAL');
    db.pragma('foreign_keys = ON');
    initializeSchema(db);
  }
  return db;
}

function initializeSchema(db: Database.Database) {
  db.exec(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      email TEXT UNIQUE NOT NULL,
      password TEXT NOT NULL,
      role TEXT NOT NULL DEFAULT 'user',
      avatar TEXT,
      bio TEXT,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS news (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      title TEXT NOT NULL,
      slug TEXT UNIQUE NOT NULL,
      excerpt TEXT,
      content TEXT NOT NULL,
      image_url TEXT,
      category TEXT NOT NULL DEFAULT 'Berita',
      tags TEXT,
      author_id INTEGER NOT NULL,
      author TEXT NOT NULL,
      status TEXT NOT NULL DEFAULT 'draft',
      is_featured INTEGER NOT NULL DEFAULT 0,
      view_count INTEGER NOT NULL DEFAULT 0,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (author_id) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS comments (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      news_id INTEGER NOT NULL,
      user_id INTEGER NOT NULL,
      user_name TEXT NOT NULL,
      content TEXT NOT NULL,
      sentiment TEXT DEFAULT 'neutral',
      sentiment_score REAL DEFAULT 0,
      is_flagged INTEGER DEFAULT 0,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (news_id) REFERENCES news(id) ON DELETE CASCADE,
      FOREIGN KEY (user_id) REFERENCES users(id)
    );

    CREATE INDEX IF NOT EXISTS idx_news_slug ON news(slug);
    CREATE INDEX IF NOT EXISTS idx_news_category ON news(category);
    CREATE INDEX IF NOT EXISTS idx_news_status ON news(status);
    CREATE INDEX IF NOT EXISTS idx_news_featured ON news(is_featured);
    CREATE INDEX IF NOT EXISTS idx_comments_news_id ON comments(news_id);
  `);

  // Seed admin user if not exists
  const adminExists = db.prepare('SELECT id FROM users WHERE email = ?').get('admin@edutech.id');
  if (!adminExists) {
    const hashedPassword = bcrypt.hashSync('admin123', 10);
    db.prepare(`
      INSERT INTO users (name, email, password, role) VALUES (?, ?, ?, ?)
    `).run('Administrator', 'admin@edutech.id', hashedPassword, 'admin');

    const adminUser = db.prepare('SELECT id FROM users WHERE email = ?').get('admin@edutech.id') as { id: number };
    seedSampleNews(db, adminUser.id);
  }
}

function seedSampleNews(db: Database.Database, authorId: number) {
  const sampleArticles = [
    {
      title: 'Tren Riset AI dalam Pendidikan Tinggi 2026: Transformasi Pembelajaran Digital',
      slug: 'tren-riset-ai-pendidikan-tinggi-2026',
      excerpt: 'Kecerdasan buatan semakin mengubah wajah pendidikan tinggi. Berbagai universitas terkemuka kini berlomba mengintegrasikan AI dalam kurikulum dan proses pembelajaran.',
      content: `Kecerdasan Buatan (AI) telah menjadi salah satu teknologi paling transformatif dalam dekade terakhir, dan dampaknya terhadap pendidikan tinggi semakin terasa signifikan pada tahun 2026 ini.

Universitas-universitas terkemuka di Indonesia, mulai dari ITB, UI, hingga ITS, kini berlomba mengintegrasikan AI dalam berbagai aspek pembelajaran. Mulai dari sistem rekomendasi kurikulum yang dipersonalisasi hingga asisten belajar berbasis AI yang tersedia 24 jam.

Dr. Ahmad Fauzi dari Departemen Ilmu Komputer UI menyatakan, "Kita sedang menyaksikan revolusi pendidikan yang sesungguhnya. AI bukan hanya alat bantu, melainkan mitra belajar yang adaptif untuk setiap mahasiswa."

Beberapa inovasi terbaru yang sedang dikembangkan meliputi sistem deteksi dini mahasiswa yang berisiko drop-out, platform pembelajaran adaptif berbasis analisis gaya belajar, serta tools penilaian otomatis yang mampu memberikan umpan balik mendalam.

Data dari Kemdikbud menunjukkan bahwa institusi yang mengadopsi AI dalam pembelajaran mengalami peningkatan rata-rata 23% dalam tingkat kelulusan tepat waktu dan 31% dalam kepuasan mahasiswa.

Namun, adopsi AI ini juga membawa tantangan tersendiri. Isu privasi data, kesenjangan infrastruktur digital, dan kebutuhan pelatihan bagi dosen menjadi fokus utama yang harus diselesaikan.`,
      image_url: 'https://images.unsplash.com/photo-1677442135703-1787eea5ce01?w=800&q=80',
      category: 'Teknologi', tags: 'AI, Pendidikan, Teknologi, Riset', is_featured: 1, status: 'published',
    },
    {
      title: 'Mahasiswa ITB Raih Juara 1 Kompetisi Robotik Internasional di Tokyo',
      slug: 'mahasiswa-itb-juara-robotik-internasional-tokyo',
      excerpt: 'Tim robotik ITB membuktikan kemampuan teknik tinggi mahasiswa Indonesia di kancah internasional dengan meraih gold medal dalam World Robot Olympiad 2026.',
      content: `Kebanggaan nasional kembali hadir dari dunia akademik. Tim robotik Institut Teknologi Bandung (ITB) berhasil merebut juara pertama dalam kompetisi World Robot Olympiad (WRO) 2026 yang diselenggarakan di Tokyo, Jepang.

Tim yang terdiri dari tiga mahasiswa Teknik Mesin dan Informatika ini mengembangkan robot bernama "GarudaBot" - sebuah robot otonom yang mampu memadamkan kebakaran di area yang tidak dapat dijangkau manusia.

"Kami menghabiskan lebih dari 18 bulan mengembangkan sistem navigasi berbasis deep learning yang dapat beradaptasi dengan lingkungan baru secara real-time," jelas Reza Pratama, ketua tim.

Robot tersebut berhasil mengungguli 47 tim dari 30 negara berbeda, menampilkan performa yang memukau para juri internasional dengan tingkat akurasi navigasi 98.7% dalam kondisi asap tebal.

Prestasi ini mendapat apresiasi langsung dari Rektor ITB dan Menteri Pendidikan, yang berjanji akan meningkatkan dukungan untuk laboratorium riset robotika di kampus-kampus teknik Indonesia.`,
      image_url: 'https://images.unsplash.com/photo-1485827404703-89b55fcc595e?w=800&q=80',
      category: 'Kampus', tags: 'Robotik, ITB, Internasional, Prestasi', is_featured: 1, status: 'published',
    },
    {
      title: 'Beasiswa LPDP 2027 Dibuka: Kuota 10.000 Penerima untuk S2 dan S3',
      slug: 'beasiswa-lpdp-2027-dibuka-kuota-10000',
      excerpt: 'Pemerintah melalui LPDP membuka pendaftaran beasiswa tahun 2027 dengan kuota terbesar sepanjang sejarah. Simak syarat dan cara daftarnya.',
      content: `Lembaga Pengelola Dana Pendidikan (LPDP) kembali membuka pendaftaran beasiswa untuk tahun akademik 2027 dengan total kuota yang mencapai 10.000 penerima - angka tertinggi dalam sejarah program beasiswa pemerintah Indonesia.

Pendaftaran dibuka mulai 1 Oktober 2026 hingga 31 Januari 2027 melalui portal resmi lpdp.kemenkeu.go.id.

Prioritas diberikan kepada bidang-bidang strategis nasional seperti teknologi informasi, energi terbarukan, pertanian modern, kesehatan, dan infrastruktur. Selain itu, kuota khusus disiapkan untuk daerah tertinggal, terdepan, dan terluar (3T).

Beberapa perubahan signifikan dari tahun sebelumnya antara lain: penghapusan batas usia maksimum untuk jalur afirmasi, penambahan program joint-degree dengan universitas riset terbaik dunia, dan peningkatan tunjangan hidup hingga 40%.`,
      image_url: 'https://images.unsplash.com/photo-1523050854058-8df90110c9f1?w=800&q=80',
      category: 'Pendidikan', tags: 'Beasiswa, LPDP, Pendidikan, S2, S3', is_featured: 0, status: 'published',
    },
    {
      title: 'Seminar Nasional SENIKO 2026: Menguatkan Resiliensi Negeri Menghadapi Ancaman Cyber',
      slug: 'seminar-nasional-seniko-2026-resiliensi-cyber',
      excerpt: 'SENIKO 2026 menghadirkan pakar keamanan siber dari dalam dan luar negeri untuk membahas strategi pertahanan digital Indonesia di era hyperconnected.',
      content: `Seminar Nasional Inovasi dan Komunikasi (SENIKO) 2026 sukses diselenggarakan di Auditorium Utama Universitas Podomoro, Jakarta, dengan tema besar "Menguatkan Resiliensi Negeri Terhadap Ancaman Cyber Warfare".

Acara yang dihadiri lebih dari 2.000 peserta dari berbagai institusi akademik dan industri ini menampilkan keynote speech dari beberapa nama terkemuka.

Panel diskusi membahas berbagai topik krusial seperti infrastruktur kritis yang rentan terhadap serangan, strategi pertahanan berlapis untuk instansi pemerintah, regulasi terbaru terkait perlindungan data pribadi, dan peran akademisi dalam mencetak talenta cybersecurity.

"Ancaman siber bukan lagi isu teknis semata - ini adalah isu keamanan nasional yang membutuhkan kolaborasi seluruh elemen bangsa," tegas Prof. Budi Rahardjo dalam sesi penutup.`,
      image_url: 'https://images.unsplash.com/photo-1550751827-4bd374c3f58b?w=800&q=80',
      category: 'Event', tags: 'SENIKO, Cybersecurity, Seminar, Keamanan', is_featured: 1, status: 'published',
    },
    {
      title: 'Lowongan Kerja: Google Indonesia Buka 500 Posisi untuk Fresh Graduate 2027',
      slug: 'lowongan-kerja-google-indonesia-500-posisi-2027',
      excerpt: 'Google Indonesia membuka rekrutmen besar-besaran untuk lulusan baru dari berbagai jurusan. Gaji kompetitif dan benefit lengkap ditawarkan.',
      content: `Google Indonesia mengumumkan program rekrutmen besar-besaran untuk tahun 2027, membuka 500 posisi bagi fresh graduate dari universitas-universitas terkemuka di Indonesia.

Posisi yang tersedia mencakup berbagai bidang mulai dari Software Engineering, Data Science, Product Management, UX Design, hingga Business Development.

Kompensasi yang ditawarkan sangat kompetitif, dengan gaji pokok mulai dari Rp 15 juta hingga Rp 35 juta per bulan untuk fresh graduate, ditambah berbagai benefit seperti asuransi kesehatan keluarga, saham perusahaan, dan program pengembangan karir intensif.

Pendaftaran dibuka hingga 28 Februari 2027 melalui careers.google.com.`,
      image_url: 'https://images.unsplash.com/photo-1573804633927-bfcbcd909acd?w=800&q=80',
      category: 'Loker', tags: 'Google, Lowongan, Fresh Graduate, Karir', is_featured: 0, status: 'published',
    },
    {
      title: 'Riset Terbaru: Metode Pembelajaran Hybrid Tingkatkan Nilai Mahasiswa 40%',
      slug: 'riset-pembelajaran-hybrid-tingkatkan-nilai-40-persen',
      excerpt: 'Penelitian longitudinal selama 3 tahun dari Universitas Gadjah Mada membuktikan bahwa metode hybrid learning secara konsisten menghasilkan output akademik lebih baik.',
      content: `Sebuah penelitian longitudinal yang dilakukan selama tiga tahun oleh tim riset dari Universitas Gadjah Mada (UGM) telah menghasilkan temuan yang mengejutkan sekaligus menggembirakan dunia pendidikan.

Studi yang melibatkan lebih dari 5.000 mahasiswa dari berbagai program studi ini menemukan bahwa model pembelajaran hybrid secara konsisten menghasilkan peningkatan nilai akademik rata-rata sebesar 40.3% dibandingkan metode konvensional.

Prof. Dr. Siti Rahayu, peneliti utama dalam studi ini, menjelaskan bahwa kunci keberhasilan hybrid learning bukan semata-mata pada teknologinya, melainkan pada desain pedagogi yang menggabungkan kelebihan kedua modalitas tersebut.

"Kami menemukan bahwa ketika sesi online dirancang untuk aktivitas low-order thinking seperti penguasaan konsep dasar, dan sesi tatap muka difokuskan pada diskusi kritis dan penerapan, hasilnya jauh lebih optimal," papar Prof. Siti.`,
      image_url: 'https://images.unsplash.com/photo-1501504905252-473c47e087f8?w=800&q=80',
      category: 'Riset', tags: 'Riset, Hybrid Learning, Pendidikan, UGM', is_featured: 0, status: 'published',
    },
  ];

  const insertNews = db.prepare(`
    INSERT INTO news (title, slug, excerpt, content, image_url, category, tags, author_id, author, is_featured, status)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);

  for (const article of sampleArticles) {
    insertNews.run(
      article.title, article.slug, article.excerpt, article.content,
      article.image_url, article.category, article.tags, authorId, 'Redaksi EduTech',
      article.is_featured, article.status
    );
  }
}
