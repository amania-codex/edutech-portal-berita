-- ========================================================
-- Script Pembersihan Data (Menyisakan Akun Admin Saja)
-- Jalankan di phpMyAdmin / HeidiSQL pada database: news_portal
-- ========================================================

SET FOREIGN_KEY_CHECKS = 0;

-- Kosongkan seluruh data artikel & komentar dummy
TRUNCATE TABLE comments;
TRUNCATE TABLE articles;

-- Kosongkan data event, pembicara, materi, dan registrasi
TRUNCATE TABLE registrations;
TRUNCATE TABLE materials;
TRUNCATE TABLE speakers;
TRUNCATE TABLE events;

-- Kosongkan data kategori, banner, rekening bank, & notifikasi
TRUNCATE TABLE categories;
TRUNCATE TABLE banners;
TRUNCATE TABLE bank_accounts;
TRUNCATE TABLE notifications;
TRUNCATE TABLE user_activities;

-- Hapus semua akun user KECUALI akun Admin
DELETE FROM users WHERE role != 'admin' OR email != 'admin@edutech.id';

SET FOREIGN_KEY_CHECKS = 1;
