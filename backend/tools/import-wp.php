<?php
/**
 * Memindahkan berita dari database WordPress lama (edukator.id) ke news_portal.
 *
 * Persiapan:
 *   1. mysql -uroot -e "CREATE DATABASE wp_edukator CHARACTER SET utf8mb4"
 *      mysql -uroot --default-character-set=utf8mb4 wp_edukator < "DATA EDUKATOR/u229455712_newedukator.sql"
 *   2. Letakkan isi wp-content/uploads (2022, 2023, 2025, 2026, ...) di public/uploads/wp/
 *
 * Jalankan dari folder backend:
 *   php tools/import-wp.php --dry-run     cek saja, tidak menulis apa pun
 *   php tools/import-wp.php               impor; yang sudah pernah diimpor dilewati
 *   php tools/import-wp.php --overwrite   timpa ulang artikel impor dengan isi WordPress
 *                                         (suntingan dari dashboard akan hilang)
 *
 * Artikel dicocokkan lewat kolom wp_id, jadi aman dijalankan berulang.
 * Artikel yang dihapus admin di sistem baru tidak dimunculkan lagi.
 * Koneksi: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, WP_DB_NAME.
 */

ini_set('memory_limit', '1G');
mb_internal_encoding('UTF-8');
mysqli_report(MYSQLI_REPORT_ERROR | MYSQLI_REPORT_STRICT);

$DRY = in_array('--dry-run', $argv, true);
$OVERWRITE = in_array('--overwrite', $argv, true);
$UPLOADS = __DIR__ . '/../public/uploads/wp';
$P = 'wp_';

function env(string $k, string $d): string { $v = getenv($k); return $v === false || $v === '' ? $d : $v; }
function connect(string $db): mysqli {
    $c = new mysqli(env('DB_HOST', '127.0.0.1'), env('DB_USER', 'root'), env('DB_PASSWORD', ''), $db, (int) env('DB_PORT', '3306'));
    $c->set_charset('utf8mb4');
    return $c;
}
function rows(mysqli $c, string $sql, array $args = []): array { return $c->execute_query($sql, $args)->fetch_all(MYSQLI_ASSOC); }
function decode(string $s): string { return trim(html_entity_decode($s, ENT_QUOTES | ENT_HTML5, 'UTF-8')); }

// ===========================================================================
// Konten
// ===========================================================================

const BLOCKS = '(?:table|thead|tfoot|caption|col|colgroup|tbody|tr|td|th|div|dl|dd|dt|ul|ol|li|pre|form|map|area|blockquote|address|style|p|h[1-6]|hr|fieldset|legend|section|article|aside|hgroup|header|footer|nav|figure|figcaption|details|menu|summary)';

/** Port ringkas wpautop(): baris kosong → <p>, baris tunggal → <br />. */
function autop(string $t): string {
    $t = trim($t);
    if ($t === '') return '';
    $B = BLOCKS;
    $t .= "\n";
    $t = preg_replace('#<br\s*/?>\s*<br\s*/?>#i', "\n\n", $t);
    $t = preg_replace("#(<$B(?=[\s/>]))#i", "\n\n$1", $t);
    $t = preg_replace("#(</$B>)#i", "$1\n\n", $t);
    $out = '';
    foreach (preg_split('#\n\s*\n#', $t, -1, PREG_SPLIT_NO_EMPTY) as $p) {
        $p = trim($p);
        if ($p !== '') $out .= "<p>$p</p>\n";
    }
    $t = preg_replace('#<p>(?:\s|&nbsp;|\x{00A0})*</p>#u', '', $out);
    $t = preg_replace("#<p>\s*(</?$B(?:\s[^>]*)?/?>)\s*</p>#i", '$1', $t);
    $t = preg_replace('#<p>(<li.+?)</p>#is', '$1', $t);
    $t = preg_replace("#<p>\s*(</?$B(?:\s[^>]*)?/?>)#i", '$1', $t);
    $t = preg_replace("#(</?$B(?:\s[^>]*)?/?>)\s*</p>#i", '$1', $t);
    // blok yang dibuka & ditutup di paragraf terpisah (mis. <p style>..\n\n..</p>) ditangani tidyP()
    $t = preg_replace('#(?<!<br />)\s*\n#', "<br />\n", $t);
    $t = preg_replace("#(</?$B(?:\s[^>]*)?/?>)\s*<br />#i", '$1', $t);
    $t = preg_replace('#<br />(\s*</?(?:p|li|div|dl|dd|dt|th|pre|td|ul|ol|h[1-6]|blockquote|figure|figcaption)[^>]*>)#i', '$1', $t);
    $t = preg_replace('#\n</p>#', '</p>', $t);
    return trim(preg_replace('#<br />\s*$#', '', $t));
}

/** Paksa HTML seimbang agar konten rusak tidak merusak layout halaman. */
function balance(string $html): string {
    $doc = new DOMDocument();
    libxml_use_internal_errors(true);
    $doc->loadHTML('<?xml encoding="UTF-8"><div id="wp-root">' . $html . '</div>', LIBXML_HTML_NODEFDTD);
    libxml_clear_errors();
    $body = $doc->getElementsByTagName('body')->item(0);
    $out = '';
    foreach ($body->childNodes as $n) {
        if ($n instanceof DOMElement && $n->getAttribute('id') === 'wp-root') {
            foreach ($n->childNodes as $c) $out .= $doc->saveHTML($c);
        } else {
            $out .= $doc->saveHTML($n); // sisa setelah </div> liar di konten
        }
    }
    return trim($out);
}

class Content {
    /** @var array<string,string> slug publik WP (aktif & lama) → slug baru */
    public array $slugs = [];
    /** @var array<string,true> file upload yang tidak ditemukan */
    public array $missing = [];
    private array $protected = [];

    public function __construct(private string $uploads, private Closure $image) {}

    private function protect(string $html): string {
        $this->protected[] = $html;
        return "\n\n<div data-wpph=\"" . (count($this->protected) - 1) . "\"></div>\n\n";
    }

    private function embed(string $url): ?string {
        $u = html_entity_decode(trim($url));
        if (preg_match('~(?:youtube\.com/(?:watch\?(?:\S*&)?v=|shorts/|live/|embed/)|youtu\.be/)([\w-]{11})~', $u, $m)) {
            return $this->protect('<div class="video-embed"><iframe src="https://www.youtube.com/embed/' . $m[1] . '" title="Video YouTube" loading="lazy" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe></div>');
        }
        return null;
    }

    private function linkify(string $u): string {
        $e = htmlspecialchars(html_entity_decode(trim($u)), ENT_QUOTES);
        return "\n\n<a href=\"$e\" target=\"_blank\" rel=\"noopener\">$e</a>\n\n";
    }

    /** Varian ukuran (-300x228) yang tidak ada di disk diganti file aslinya. */
    public function upload(string $rel): string {
        $exists = fn(string $p) => is_file($this->uploads . '/' . rawurldecode($p));
        $rel = preg_replace('#[?\#].*$#', '', $rel);
        if ($exists($rel)) return $rel;
        $orig = preg_replace('#-\d+x\d+(\.\w+)$#', '$1', $rel);
        if ($orig !== $rel && $exists($orig)) return $orig;
        $this->missing[$rel] = true;
        return $rel;
    }

    private function siteUrl(array $m): string {
        $path = $m[1] ?? '';
        if (str_starts_with($path, '/wp-content/uploads/')) return '/uploads/wp/' . $this->upload(substr($path, 20));
        $segs = array_values(array_filter(explode('/', preg_replace('#[?\#].*$#', '', $path)), 'strlen'));
        if (!$segs) return '/';
        $last = end($segs);
        if ($segs[0] === 'category' && count($segs) > 1) return "/kategori/$last";
        if ($segs[0] === 'tag' && count($segs) > 1) return '/cari?tag=' . rawurlencode(str_replace('-', ' ', rawurldecode($last)));
        return isset($this->slugs[$last]) ? '/berita/' . $this->slugs[$last] : $m[0];
    }

    public function transform(string $raw): string {
        $this->protected = [];
        $s = str_replace(["\r\n", "\r"], "\n", $raw);
        $gutenberg = str_contains($s, '<!-- wp:');
        $s = preg_replace('#<!--.*?-->#s', '', $s);
        $s = preg_replace('#<(script|style)\b.*?</\1>#is', '', $s);
        // sisa copy-paste: data-*, class utilitas, id/dir/tabindex, dan event handler inline
        $s = preg_replace('#(?<=[\s"\'])(?:data-[\w-]+|dir|tabindex|on[a-z]+)="[^"]*"#i', '', $s);
        $s = preg_replace('#\s+class="[^"]*[\[\]:][^"]*"#', '', $s);

        // hanya di dalam atribut; URL yang tertulis sebagai teks dibiarkan apa adanya
        $s = preg_replace('#\s(?:srcset|sizes)="[^"]*"#i', '', $s);
        $s = preg_replace_callback('#\b((?:href|src)\s*=\s*["\'])(https?://(?:www\.)?edukator\.id(?![\w.-])(/[^"\'<>]*)?)#i',
            fn($m) => $m[1] . $this->siteUrl([$m[2], trim($m[3] ?? '')]), $s);

        // Foto disisipkan sebagai thumbnail (mis. -300x200) lalu ditarik lebar; WordPress menutupinya
        // dengan srcset. Di sini langsung dipakai versi besarnya.
        $s = preg_replace_callback('#<img\b[^>]*>#i', function ($m) {
            if (!preg_match('#wp-image-(\d+)#', $m[0], $id) || !preg_match('#src="/uploads/wp/([^"]+)"#', $m[0], $src)) return $m[0];
            $stem = fn(string $f) => preg_replace('#(-scaled)?(-\d+x\d+)?\.\w+$#', '', $f);
            $big = ($this->image)((int) $id[1]);
            return $big !== '' && $stem($big) === '/uploads/wp/' . $stem($src[1]) ? str_replace($src[0], "src=\"$big\"", $m[0]) : $m[0];
        }, $s);

        $s = preg_replace_callback('#\[caption[^\]]*\](.*?)\[/caption\]#s', function ($m) {
            if (!preg_match('#^\s*((?:<a\b[^>]*>\s*)?<img\b[^>]*>(?:\s*</a>)?)(.*)$#s', $m[1], $x)) return $m[1];
            $cap = trim($x[2]);
            return $this->protect("<figure class=\"wp-caption\">$x[1]" . ($cap !== '' ? "<figcaption>$cap</figcaption>" : '') . '</figure>');
        }, $s);

        $s = preg_replace_callback('#\[embed[^\]]*\](.*?)\[/embed\]#s', fn($m) => $this->embed($m[1]) ?? $this->linkify($m[1]), $s);
        $s = preg_replace_callback('#<div class="wp-block-embed__wrapper">\s*(https?://[^\s<]+)\s*</div>#', fn($m) => $this->embed($m[1]) ?? $this->linkify($m[1]), $s);
        $s = preg_replace_callback('#^[ \t]*(https?://[^\s<>"]+)[ \t]*$#m', fn($m) => $this->embed($m[1]) ?? $this->linkify(rtrim($m[1], ',.')), $s);

        foreach (['pre', 'table', 'iframe', 'svg'] as $tag) {
            $s = preg_replace_callback("#<$tag\\b.*?</$tag>#is", fn($m) => $this->protect($m[0]), $s);
        }
        if (!$gutenberg) $s = autop($s);
        for ($i = 0; $i < 5 && str_contains($s, 'data-wpph'); $i++) {
            $s = preg_replace_callback('#<div data-wpph="(\d+)"></div>#', fn($m) => $this->protected[(int) $m[1]], $s);
        }
        if ($s === null) throw new RuntimeException('regex gagal: ' . preg_last_error_msg());
        return trim($s) === '' ? '' : balance($s);
    }
}

function plain(string $html): string {
    $t = html_entity_decode(preg_replace('#<[^>]*>#', ' ', $html), ENT_QUOTES | ENT_HTML5, 'UTF-8');
    return trim(preg_replace('#[\s\x{00A0}]+#u', ' ', $t));
}
function excerpt(string $html, int $max = 160): string {
    $t = plain($html);
    if (mb_strlen($t) <= $max) return $t;
    $cut = mb_substr($t, 0, $max);
    $sp = mb_strrpos($cut, ' ');
    if ($sp > $max / 2) $cut = mb_substr($cut, 0, $sp);
    return rtrim($cut, ' ,.;:-') . '...';
}
function cleanSlug(string $s): string {
    return trim(preg_replace('#[^a-z0-9]+#', '-', mb_strtolower(rawurldecode($s))), '-');
}

// ===========================================================================
// Struktur tabel (identik dengan hasil AutoMigrate model Go)
// ===========================================================================

function migrate(mysqli $d): void {
    $cols = array_column(rows($d, "SELECT COLUMN_NAME c FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='articles'"), 'c');
    if (!$cols) exit("Tabel articles belum ada. Jalankan backend sekali dulu agar AutoMigrate membuat tabel dasar.\n");
    foreach (['category_slug' => 'varchar(150)', 'read_time' => "bigint DEFAULT '5'", 'wp_id' => "bigint unsigned DEFAULT '0'"] as $c => $def) {
        if (!in_array($c, $cols, true)) $d->query("ALTER TABLE articles ADD COLUMN `$c` $def");
    }
    $d->query("UPDATE articles SET is_published = (status = 'published') WHERE is_published IS NULL");
    $d->query('ALTER TABLE articles MODIFY is_published tinyint(1) NOT NULL');
    $idx = array_column(rows($d, "SELECT DISTINCT INDEX_NAME i FROM information_schema.STATISTICS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='articles'"), 'i');
    foreach (['category', 'category_slug', 'is_featured', 'is_trending', 'is_published', 'status', 'wp_id'] as $c) {
        if (!in_array("idx_articles_$c", $idx, true)) $d->query("CREATE INDEX idx_articles_$c ON articles(`$c`)");
    }
    $opt = 'ENGINE=InnoDB DEFAULT CHARSET=utf8mb4';
    $d->query("CREATE TABLE IF NOT EXISTS categories (id bigint unsigned NOT NULL AUTO_INCREMENT, created_at datetime(3) NULL, updated_at datetime(3) NULL, deleted_at datetime(3) NULL,
        name varchar(100) NOT NULL, slug varchar(150) NOT NULL, parent_id bigint unsigned NULL, wp_id bigint unsigned DEFAULT '0', PRIMARY KEY (id),
        UNIQUE KEY uni_categories_name (name), UNIQUE KEY uni_categories_slug (slug), KEY idx_categories_deleted_at (deleted_at), KEY idx_categories_parent_id (parent_id), KEY idx_categories_wp_id (wp_id)) $opt");
    $cc = array_column(rows($d, "SELECT COLUMN_NAME c FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='categories'"), 'c');
    if (!in_array('parent_id', $cc, true)) $d->query('ALTER TABLE categories ADD COLUMN parent_id bigint unsigned NULL, ADD KEY idx_categories_parent_id (parent_id)');
    if (!in_array('wp_id', $cc, true)) $d->query("ALTER TABLE categories ADD COLUMN wp_id bigint unsigned DEFAULT '0', ADD KEY idx_categories_wp_id (wp_id)");
    $d->query("CREATE TABLE IF NOT EXISTS article_categories (article_id bigint unsigned NOT NULL, category_id bigint unsigned NOT NULL, PRIMARY KEY (article_id, category_id), KEY idx_article_categories_category_id (category_id)) $opt");
    $d->query("CREATE TABLE IF NOT EXISTS article_redirects (id bigint unsigned NOT NULL AUTO_INCREMENT, old_slug varchar(255) NOT NULL, article_id bigint unsigned NOT NULL, PRIMARY KEY (id),
        UNIQUE KEY idx_article_redirects_old_slug (old_slug), KEY idx_article_redirects_article_id (article_id)) $opt");
}

// ===========================================================================
// Baca sumber
// ===========================================================================

$wp = connect(env('WP_DB_NAME', 'wp_edukator'));
$db = connect(env('DB_NAME', 'news_portal'));
if (!is_dir($UPLOADS)) exit("Folder $UPLOADS tidak ditemukan.\n");

$terms = rows($wp, "SELECT t.term_id, tt.term_taxonomy_id ttid, t.name, t.slug, tt.taxonomy, tt.parent FROM {$P}terms t JOIN {$P}term_taxonomy tt ON tt.term_id=t.term_id WHERE tt.taxonomy IN ('category','post_tag')");
$cats = []; $catByTt = []; $tagByTt = [];
foreach ($terms as $t) {
    if ($t['taxonomy'] === 'category') {
        $cats[$t['term_id']] = ['name' => decode($t['name']), 'slug' => $t['slug'], 'parent' => (int) $t['parent']];
        $catByTt[$t['ttid']] = (int) $t['term_id'];
    } else {
        $tagByTt[$t['ttid']] = str_replace(',', ' ', decode($t['name']));
    }
}
$postCats = []; $postTags = [];
foreach (rows($wp, "SELECT object_id o, term_taxonomy_id tt FROM {$P}term_relationships ORDER BY term_order, term_taxonomy_id") as $r) {
    if (isset($catByTt[$r['tt']])) $postCats[$r['o']][] = $catByTt[$r['tt']];
    elseif (isset($tagByTt[$r['tt']])) $postTags[$r['o']][] = $tagByTt[$r['tt']];
}
$meta = []; $oldSlugs = [];
foreach (rows($wp, "SELECT post_id p, meta_key k, meta_value v FROM {$P}postmeta WHERE meta_key IN ('_thumbnail_id','_wp_attached_file','_wp_attachment_metadata','_yoast_wpseo_primary_category','_yoast_wpseo_estimated-reading-time-minutes','_wp_old_slug','post_views_count')") as $m) {
    if ($m['k'] === '_wp_old_slug') $oldSlugs[$m['p']][] = $m['v'];
    else $meta[$m['p']][$m['k']] = $m['v'];
}
$views = array_column(rows($wp, "SELECT postid, pageviews FROM {$P}popularpostsdata"), 'pageviews', 'postid');
$authors = array_map('decode', array_column(rows($wp, "SELECT ID, display_name FROM {$P}users"), 'display_name', 'ID'));
$sticky = unserialize(rows($wp, "SELECT option_value v FROM {$P}options WHERE option_name='sticky_posts'")[0]['v'] ?? 'a:0:{}', ['allowed_classes' => false]) ?: [];
$posts = rows($wp, "SELECT ID, post_author, CAST(post_date AS CHAR) d, CAST(post_modified AS CHAR) m, post_title, post_content, post_excerpt, post_status, post_name FROM {$P}posts WHERE post_type='post' AND post_status IN ('publish','draft','future','pending','private') ORDER BY post_date, ID");

// ===========================================================================
// Tulis
// ===========================================================================

if (!$DRY) migrate($db);
$hasWp = (bool) rows($db, "SELECT 1 FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='articles' AND COLUMN_NAME='wp_id'");
$existing = []; $taken = [];
foreach (rows($db, 'SELECT id, slug, view_count, is_featured, deleted_at' . ($hasWp ? ', wp_id' : ', 0 wp_id') . ' FROM articles') as $a) {
    if ($a['wp_id']) $existing[$a['wp_id']] = $a;
    else $taken[$a['slug']] = true;
}

$image = function (int $att) use ($meta, $UPLOADS): string {
    $file = $meta[$att]['_wp_attached_file'] ?? '';
    if ($file === '') return '';
    $info = @unserialize($meta[$att]['_wp_attachment_metadata'] ?? '', ['allowed_classes' => false]);
    $tries = [];
    foreach (['large', 'medium_large'] as $size) // ±1024px: cukup tajam, jauh lebih ringan dari file asli
        if (!empty($info['sizes'][$size]['file'])) $tries[] = dirname($file) . '/' . $info['sizes'][$size]['file'];
    $tries[] = $file;
    foreach ($tries as $t) if (is_file("$UPLOADS/$t")) return "/uploads/wp/$t";
    return '';
};

$content = new Content($UPLOADS, $image);
$final = [];
foreach ($posts as $p) {
    $slug = cleanSlug($p['post_name']) ?: 'draft-wp-' . $p['ID'];
    if (isset($taken[$slug])) $slug .= '-' . $p['ID'];
    $taken[$slug] = true;
    $final[$p['ID']] = $slug;
}
foreach ($posts as $p) if ($p['post_status'] === 'publish') foreach ($oldSlugs[$p['ID']] ?? [] as $o) if ($o !== '') $content->slugs[$o] = $final[$p['ID']];
foreach ($posts as $p) if ($p['post_status'] === 'publish' && $p['post_name'] !== '') $content->slugs[$p['post_name']] = $final[$p['ID']];

/** Kategori badge: pilihan Yoast, kecuali terlalu umum → kategori paling spesifik. */
$mainCat = function (array $ids, int $primary) use ($cats): ?int {
    $depth = function (int $id) use ($cats) { $n = 0; while (($id = $cats[$id]['parent'] ?? 0) && $n < 9) $n++; return $n; };
    $generic = fn(int $id) => in_array($cats[$id]['slug'], ['berita', 'umum'], true);
    if ($primary && isset($cats[$primary]) && in_array($primary, $ids, true) && !$generic($primary)) return $primary;
    if (!$ids) return isset($cats[$primary]) ? $primary : null;
    usort($ids, fn($a, $b) => [$generic($a), -$depth($a)] <=> [$generic($b), -$depth($b)]);
    return $ids[0];
};

$stat = ['baru' => 0, 'ditimpa' => 0, 'dilewati' => 0, 'tanpa_gambar' => 0];
$noImage = [];
$db->begin_transaction();
try {
    $catId = [];
    if (!$DRY) {
        foreach ($cats as $term => $c) {
            $f = rows($db, 'SELECT id FROM categories WHERE wp_id=? OR slug=? OR name=? ORDER BY wp_id=? DESC LIMIT 1', [$term, $c['slug'], $c['name'], $term]);
            if ($f) {
                $catId[$term] = (int) $f[0]['id'];
                $db->execute_query('UPDATE categories SET wp_id=? WHERE id=?', [$term, $catId[$term]]);
            } else {
                $db->execute_query('INSERT INTO categories (created_at,updated_at,name,slug,wp_id) VALUES (NOW(3),NOW(3),?,?,?)', [$c['name'], $c['slug'], $term]);
                $catId[$term] = $db->insert_id;
            }
        }
        foreach ($cats as $term => $c) if ($c['parent'] && isset($catId[$c['parent']]))
            $db->execute_query('UPDATE categories SET parent_id=? WHERE id=?', [$catId[$c['parent']], $catId[$term]]);
    }

    foreach ($posts as $p) {
        $id = (int) $p['ID'];
        $old = $existing[$id] ?? null;
        if ($old && ($old['deleted_at'] !== null || !$OVERWRITE)) { $stat['dilewati']++; continue; }

        $title = decode(strip_tags($p['post_title']));
        $raw = $p['post_content'];
        if (mb_strlen($title) > 200) { // seluruh naskah tertempel di kolom judul
            if (trim($raw) === '') $raw = preg_replace('#\s{2,}#', "

", $p['post_title']);
            $title = excerpt($title, 120);
        }
        $html = $content->transform($raw);
        $img = $image((int) ($meta[$id]['_thumbnail_id'] ?? 0));
        if ($img === '' && preg_match('#<img\b[^>]*\bsrc="(/uploads/wp/[^"]+)"#', $html, $m)) $img = $m[1];
        if ($img === '') { $stat['tanpa_gambar']++; $noImage[] = "$id ({$p['post_status']})"; }

        $ids = array_values(array_unique($postCats[$id] ?? []));
        $mc = $mainCat($ids, (int) ($meta[$id]['_yoast_wpseo_primary_category'] ?? 0));
        $pub = $p['post_status'] === 'publish';
        $words = count(preg_split('#\s+#', plain($html), -1, PREG_SPLIT_NO_EMPTY));
        $row = [
            'created_at' => $p['d'], 'updated_at' => $p['m'] > '1971' ? $p['m'] : $p['d'],
            'title' => $title ?: '(Tanpa judul)', 'slug' => $final[$id],
            'excerpt' => decode($p['post_excerpt']) ?: excerpt($html), 'content' => $html, 'image_url' => $img,
            'category' => $mc ? $cats[$mc]['name'] : '', 'category_slug' => $mc ? $cats[$mc]['slug'] : '',
            'author' => $authors[$p['post_author']] ?? 'Redaksi', 'author_id' => 0,
            'tags' => mb_substr(implode(', ', $postTags[$id] ?? []), 0, 500),
            'view_count' => max((int) ($old['view_count'] ?? 0), (int) ($views[$id] ?? 0), (int) ($meta[$id]['post_views_count'] ?? 0)),
            'read_time' => (int) ($meta[$id]['_yoast_wpseo_estimated-reading-time-minutes'] ?? 0) ?: max(1, (int) ceil($words / 200)),
            'is_featured' => (int) (($old['is_featured'] ?? 0) || in_array($id, $sticky, true)),
            'is_trending' => 0, 'is_published' => (int) $pub, 'status' => $pub ? 'published' : 'draft', 'wp_id' => $id,
        ];
        $stat[$old ? 'ditimpa' : 'baru']++;
        if ($DRY) continue;

        if ($old) {
            unset($row['is_trending']);
            $db->execute_query('UPDATE articles SET ' . implode(',', array_map(fn($k) => "`$k`=?", array_keys($row))) . ' WHERE id=?', [...array_values($row), $old['id']]);
            $aid = (int) $old['id'];
        } else {
            $db->execute_query('INSERT INTO articles (`' . implode('`,`', array_keys($row)) . '`) VALUES (' . implode(',', array_fill(0, count($row), '?')) . ')', array_values($row));
            $aid = $db->insert_id;
        }
        $db->execute_query('DELETE FROM article_categories WHERE article_id=?', [$aid]);
        foreach ($ids as $term) $db->execute_query('INSERT INTO article_categories VALUES (?,?)', [$aid, $catId[$term]]);
        if ($pub) foreach (array_unique([$p['post_name'], ...($oldSlugs[$id] ?? [])]) as $o) {
            if ($o !== '' && $o !== $final[$id] && ($content->slugs[$o] ?? '') === $final[$id])
                $db->execute_query('INSERT INTO article_redirects (old_slug, article_id) VALUES (?,?) ON DUPLICATE KEY UPDATE article_id=VALUES(article_id)', [$o, $aid]);
        }
    }
    // artikel buatan sistem baru belum punya category_slug
    if (!$DRY) $db->query("UPDATE articles a JOIN categories c ON c.name=a.category SET a.category_slug=c.slug WHERE COALESCE(a.category_slug,'')=''");
    $db->commit();
} catch (Throwable $e) {
    $db->rollback();
    fwrite(STDERR, 'GAGAL, semua perubahan data dibatalkan: ' . $e->getMessage() . "\n");
    exit(1);
}

echo ($DRY ? '[DRY-RUN] ' : '') . 'Post WordPress: ' . count($posts) . ' | ' . urldecode(http_build_query($stat, '', ' | ')) . "\n";
echo 'Kategori: ' . count($cats) . ', tag: ' . count($tagByTt) . "\n";
echo 'File konten tidak ditemukan: ' . (implode(', ', array_keys($content->missing)) ?: '-') . "\n";
echo 'Tanpa gambar: ' . (implode('; ', $noImage) ?: '-') . "\n";
