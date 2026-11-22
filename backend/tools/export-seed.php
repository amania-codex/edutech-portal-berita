<?php
/**
 * Membuat backend/database/seed/berita.json.gz dari database news_portal.
 * Isinya ditanam ke backend dan dimuat otomatis saat database masih kosong
 * (lihat database/news_seeder.go).
 *
 * HANYA berita terbit, kategori, dan slug lama. Draft, pengguna, komentar, dan
 * data pribadi lain tidak pernah diekspor karena repo ini publik.
 *
 *   php tools/export-seed.php            (dijalankan dari folder backend)
 */
mysqli_report(MYSQLI_REPORT_ERROR | MYSQLI_REPORT_STRICT);
ini_set('memory_limit', '1G');
$env = fn(string $k, string $d) => getenv($k) ?: $d;
$db = new mysqli($env('DB_HOST', '127.0.0.1'), $env('DB_USER', 'root'), $env('DB_PASSWORD', ''), $env('DB_NAME', 'news_portal'), (int) $env('DB_PORT', '3306'));
$db->set_charset('utf8mb4');
$rows = fn(string $sql) => $db->query($sql)->fetch_all(MYSQLI_ASSOC);

// Foto surat polisi yang memuat NIK, telepon, dan alamat warga: tidak boleh masuk repo publik.
$private = ['/uploads/wp/2025/12/stnk-bluur-683x1024.webp', '/uploads/wp/2026/01/STNK-TIWI--796x1024.webp'];
$redact = function (array $a) use ($private): array {
    foreach ($private as $file) {
        if ($a['image_url'] === $file) $a['image_url'] = '';
        $a['content'] = preg_replace('#<img\b[^>]*src="' . preg_quote($file, '#') . '"[^>]*>#', '<em>Foto surat keterangan tidak ditampilkan untuk melindungi data pribadi.</em>', $a['content']);
    }
    return $a;
};

$cats = [];
foreach ($rows('SELECT c.name, c.slug, p.slug parent, c.wp_id FROM categories c LEFT JOIN categories p ON p.id=c.parent_id WHERE c.deleted_at IS NULL ORDER BY c.parent_id IS NOT NULL, c.id') as $c)
    $cats[] = ['name' => $c['name'], 'slug' => $c['slug'], 'parent' => $c['parent'] ?? '', 'wpId' => (int) $c['wp_id']];
$artCats = $redirects = [];
foreach ($rows('SELECT ac.article_id a, c.slug FROM article_categories ac JOIN categories c ON c.id=ac.category_id AND c.deleted_at IS NULL ORDER BY ac.category_id') as $r) $artCats[$r['a']][] = $r['slug'];
foreach ($rows('SELECT article_id a, old_slug FROM article_redirects ORDER BY id') as $r) $redirects[$r['a']][] = $r['old_slug'];

$articles = [];
foreach ($rows('SELECT * FROM articles WHERE is_published=1 AND deleted_at IS NULL ORDER BY created_at, id') as $a) {
    $a = $redact($a);
    $articles[] = [
        'wpId' => (int) $a['wp_id'], 'createdAt' => substr($a['created_at'], 0, 19), 'updatedAt' => substr($a['updated_at'], 0, 19),
        'title' => $a['title'], 'slug' => $a['slug'], 'excerpt' => $a['excerpt'], 'content' => $a['content'], 'imageUrl' => $a['image_url'],
        'category' => $a['category'], 'categorySlug' => $a['category_slug'] ?? '', 'author' => $a['author'], 'tags' => $a['tags'],
        'viewCount' => (int) $a['view_count'], 'readTime' => (int) $a['read_time'], 'isFeatured' => (bool) $a['is_featured'],
        'categories' => $artCats[$a['id']] ?? [], 'oldSlugs' => $redirects[$a['id']] ?? [],
    ];
}
$json = json_encode(['categories' => $cats, 'articles' => $articles], JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES | JSON_THROW_ON_ERROR);
$out = __DIR__ . '/../database/seed/berita.json.gz';
@mkdir(dirname($out), 0777, true);
file_put_contents($out, gzencode($json, 9));
printf("%d berita, %d kategori, %d slug lama → %.1f MB (asli %.1f MB)\n", count($articles), count($cats), array_sum(array_map('count', $redirects)), filesize($out) / 1048576, strlen($json) / 1048576);
