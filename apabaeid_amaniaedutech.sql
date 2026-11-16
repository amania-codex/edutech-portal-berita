-- phpMyAdmin SQL Dump
-- version 5.2.3
-- https://www.phpmyadmin.net/
--
-- Host: localhost:3306
-- Generation Time: Oct 25, 2026 at 09:59 PM
-- Server version: 10.6.28-MariaDB
-- PHP Version: 8.4.27

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `apabaeid_amaniaedutech`
--

-- --------------------------------------------------------

--
-- Table structure for table `articles`
--

CREATE TABLE `articles` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `title` varchar(255) NOT NULL,
  `slug` varchar(255) NOT NULL,
  `category` varchar(255) NOT NULL DEFAULT 'Umum',
  `image` varchar(255) DEFAULT NULL,
  `content` longtext NOT NULL,
  `read_time` int(11) NOT NULL DEFAULT 5,
  `is_published` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `bank_accounts`
--

CREATE TABLE `bank_accounts` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `bank_name` varchar(255) NOT NULL,
  `account_number` varchar(255) NOT NULL,
  `account_holder` varchar(255) NOT NULL,
  `logo` varchar(255) DEFAULT NULL,
  `is_active` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `bank_accounts`
--

INSERT INTO `bank_accounts` (`id`, `bank_name`, `account_number`, `account_holder`, `logo`, `is_active`, `created_at`, `updated_at`) VALUES
(1, 'BCA', '12424124', 'PT Edutech', 'bank-logos/QimyaBuE6x90LvZFYDtoCdq94wMFtPhPJ34cl64T.png', 1, '2026-02-24 21:58:12', '2026-02-24 21:58:12');

-- --------------------------------------------------------

--
-- Table structure for table `bank_account_event`
--

CREATE TABLE `bank_account_event` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `event_id` bigint(20) UNSIGNED NOT NULL,
  `bank_account_id` bigint(20) UNSIGNED NOT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `cache`
--

CREATE TABLE `cache` (
  `key` varchar(255) NOT NULL,
  `value` mediumtext NOT NULL,
  `expiration` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `cache_locks`
--

CREATE TABLE `cache_locks` (
  `key` varchar(255) NOT NULL,
  `owner` varchar(255) NOT NULL,
  `expiration` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `events`
--

CREATE TABLE `events` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `title` varchar(255) NOT NULL,
  `description` text NOT NULL,
  `venue` varchar(255) NOT NULL,
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  `quota` int(11) NOT NULL,
  `price` int(11) NOT NULL DEFAULT 0,
  `premium_price` int(11) DEFAULT NULL,
  `certificate_link` varchar(255) DEFAULT NULL,
  `image` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `failed_jobs`
--

CREATE TABLE `failed_jobs` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `uuid` varchar(255) NOT NULL,
  `connection` text NOT NULL,
  `queue` text NOT NULL,
  `payload` longtext NOT NULL,
  `exception` longtext NOT NULL,
  `failed_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `jobs`
--

CREATE TABLE `jobs` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `queue` varchar(255) NOT NULL,
  `payload` longtext NOT NULL,
  `attempts` tinyint(3) UNSIGNED NOT NULL,
  `reserved_at` int(10) UNSIGNED DEFAULT NULL,
  `available_at` int(10) UNSIGNED NOT NULL,
  `created_at` int(10) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `job_batches`
--

CREATE TABLE `job_batches` (
  `id` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `total_jobs` int(11) NOT NULL,
  `pending_jobs` int(11) NOT NULL,
  `failed_jobs` int(11) NOT NULL,
  `failed_job_ids` longtext NOT NULL,
  `options` mediumtext DEFAULT NULL,
  `cancelled_at` int(11) DEFAULT NULL,
  `created_at` int(11) NOT NULL,
  `finished_at` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `materials`
--

CREATE TABLE `materials` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `event_id` bigint(20) UNSIGNED NOT NULL,
  `title` varchar(255) NOT NULL,
  `type` enum('file','video') NOT NULL,
  `file_path` varchar(255) DEFAULT NULL,
  `link` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `migrations`
--

CREATE TABLE `migrations` (
  `id` int(10) UNSIGNED NOT NULL,
  `migration` varchar(255) NOT NULL,
  `batch` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `migrations`
--

INSERT INTO `migrations` (`id`, `migration`, `batch`) VALUES
(1, '0001_01_01_000000_create_users_table', 1),
(2, '0001_01_01_000001_create_cache_table', 1),
(3, '0001_01_01_000002_create_jobs_table', 1),
(4, '2025_12_18_020433_create_events_table', 1),
(5, '2025_12_18_020435_create_registrations_table', 1),
(6, '2025_12_20_230148_create_materials_table', 1),
(7, '2025_12_22_015404_add_freemium_columns_to_events_and_registrations', 1),
(8, '2025_12_22_024820_add_user_id_to_registrations_table', 1),
(9, '2025_12_29_141315_create_bank_accounts_table', 1),
(10, '2025_12_30_092919_create_bank_account_event_table', 1),
(11, '2025_12_30_113359_add_total_amount_to_registrations_table', 1),
(12, '2026_02_25_144222_add_ticket_code_to_registrations_table', 2),
(13, '2026_02_26_013955_add_certificate_link_to_events_table', 3),
(14, '2026_02_27_022653_create_articles_table', 4),
(15, '2026_02_28_032348_drop_speaker_column_from_events_table', 5),
(16, '2026_02_28_032537_create_speakers_table', 6);

-- --------------------------------------------------------

--
-- Table structure for table `password_reset_tokens`
--

CREATE TABLE `password_reset_tokens` (
  `email` varchar(255) NOT NULL,
  `token` varchar(255) NOT NULL,
  `created_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `registrations`
--

CREATE TABLE `registrations` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `ticket_code` varchar(255) DEFAULT NULL,
  `user_id` bigint(20) UNSIGNED DEFAULT NULL,
  `event_id` bigint(20) UNSIGNED NOT NULL,
  `name` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `payment_proof` varchar(255) DEFAULT NULL,
  `status` enum('pending','waiting','verified','rejected') NOT NULL DEFAULT 'pending',
  `tier` enum('free','premium') NOT NULL DEFAULT 'free',
  `total_amount` decimal(15,2) NOT NULL DEFAULT 0.00,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `sessions`
--

CREATE TABLE `sessions` (
  `id` varchar(255) NOT NULL,
  `user_id` bigint(20) UNSIGNED DEFAULT NULL,
  `ip_address` varchar(45) DEFAULT NULL,
  `user_agent` text DEFAULT NULL,
  `payload` longtext NOT NULL,
  `last_activity` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `sessions`
--

INSERT INTO `sessions` (`id`, `user_id`, `ip_address`, `user_agent`, `payload`, `last_activity`) VALUES
('2ZTPppqwxWiZvcB2hFPUYqYx0RchggQWZ53oo2VD', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiZWtSWDR3MWVLbEN4YzI4UHpjMFZSSURCM1Y4ZHRmSVY3em9TelN0YSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6Mjc6Imh0dHBzOi8vYW1hbmlhLmlkLmFwYWJhZS5pZCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773892927),
('3isFiW68DI9CJFLbq2FoXcU1GCqdb4NgjxervNMk', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiYzhkOE92dG5nVmRzOTBvUVB2eU9GVWdGbGFnT1dDV1FPTDNrQ3FxMyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzI6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9cXVhbGl0eSUyMHJ4IjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773899215),
('4TBqTbCoV4eFjAAfWHVULCSEHWkh0lnlWFk9ehSh', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiNjRkRnIwZ1N2RnlLSDRpT0NqWXRJN2pubG5jeXlEM2M3U014cUpINSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6Mjc6Imh0dHBzOi8vYW1hbmlhLmlkLmFwYWJhZS5pZCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773894750),
('4zZOLACDMUw7alRGGIXtp1dKQYqvtRepHkPvP8aI', NULL, '51.68.236.73', 'Mozilla/5.0 (compatible; MJ12bot/v2.0.5; http://mj12bot.com/)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiR2pnRGF3MkNSUXpBbDYzbkVGOElLOWliTXZXZWhLdkhYYjQ3ZWZ1USI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTY6Imh0dHA6Ly9hbWFuaWEuaWQiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773892027),
('5sG37z1ol3YfD6Fh9V1Vw2KzBzrDX6OClo8h2Hue', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoicE1JdGIwenZDYVViVDBYb0Y2R0Q0b1EyTkQ0ZVFTU01iM0ZGTWdTdyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTc6Imh0dHBzOi8vYW1hbmlhLmlkIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773894737),
('5uGWe2q7UeJg01e5M6B7hLCRP5x9sKIBN1IU20vq', NULL, '114.119.138.170', 'Mozilla/5.0 (Linux; Android 7.0;) AppleWebKit/537.36 (KHTML, like Gecko) Mobile Safari/537.36 (compatible; PetalBot;+https://webmaster.petalsearch.com/site/petalbot)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiQjdGUkhYWk9IbE1XSVpDTnJQVlpiRkRmZHJ1cVlCaUwwc1U2VUZPZCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MjU6Imh0dHA6Ly9hbWFuaWEuaWQvYXJ0aWNsZXMiO3M6NToicm91dGUiO3M6MTQ6ImFydGljbGVzLmluZGV4Ijt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773893386),
('79hpLfbuUN3clVFL6Bqchn7TbV4KP6cqPG8nkTQ3', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiVlplRnRJbnJoVDU0VUxSS2hOcGNJNFR0SkM0ZnBMekFsbXp6VUFzWSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzI6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9dGFkYWNpcCUyMDIwIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773891555),
('79kSmgz3H80JSsQz1rBmBUkGhNTlwIF1gZVCetsv', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiZE5sRERBZ1owbnhlUVZHTURTQ3FIRDg3M21hdTlGM2VqRE93cXQ1QiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NTE6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9d2hhdCUyMGlzJTIwY2lhbGlzJTIwdXNlZCUyMGZvciI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773896699),
('7CgxDKeEx1703azzM9FABRXVsTASmiBaVsOIuPtx', NULL, '45.237.151.50', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiOXFNV1pRQWw1UWRsb2tqSEQyMkg5cjlHcVlRdnFBV3d1WUh1azZzdSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9Z2ElMkZZNzE4MTg0NC5odG1sIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773891334),
('7CHZJVBqds76WIB7JUEvPR6jDA2WzNmFNR9zO3Vr', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiUmRtMjRqbndURzZFMUJLMnduVmhsTTN6VmtOMERzZHZrNTlaUVRHQSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDM6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9eGFuYXglMjBhbmQlMjBwcmVnbmFuY3kiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773899892),
('81aZxt6vxFsedVMMRSwux4zj7X2NNbEro99GLVys', NULL, '167.86.185.48', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiaXMxMm9DODd2RlVxRzBKT0VTMmFOUVF4d3Awbk12Qkk4cHppQVViMyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9Y3IlMkZSNTg4NTE0OS5odG1sIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773901841),
('8jLvBDNmPuhIUCTXXUMPvDnpnZ4kN156QNhb43Ua', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiRld6SjNXWEtFSGRqbVR5amJhM0ZQeER0a2ppNUJJV1FWQ05INGRSayI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NTI6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9YnV5JTIwc3RyYXR0ZXJhJTIwb25saW5lJTIwZmVkZXgiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773899137),
('9wFe8ciLS2n9Vm3mxMBFqbo5eSqzMpWp3bYXnaY8', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiODk4dDNST2JnZVpqa0U3NjFLRDVFd3ZBTzZkaWlYVlN6bWt2NXZnNCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzE6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9Y2lhbGlzJTIwNWciO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773901840),
('AASojomFWqUyNNxjBvwqL2Z8Ab4KyAqdW3wUc2VY', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoianF1ZTJxbGZnbDJjYkFDQXc4UTJDQnBnVU1XdXNCRk5zZnBiR0puaiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6Mjc6Imh0dHBzOi8vYW1hbmlhLmlkLmFwYWJhZS5pZCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773898342),
('AFVsHc9SPGl4cX6dEt72GPp6MekqHz0p6889svi8', NULL, '185.191.171.10', 'Mozilla/5.0 (compatible; SemrushBot/7~bl; +http://www.semrush.com/bot.html)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiWGtKSm5FSkdONE4xTkh1S2xCTTROd0s4VjVNZmtoRnM2cGdidjlEVyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTE0OiJodHRwczovL2FtYW5pYS5pZC8/cz0lRjAlOUYlOTElODklMjBhY2M2LnRvcCUyMCVGMCU5RiU5MSU4OCVGMCU5RiU4RiVCQiUyMGNvbXByYXIlMjB1bWElMjBUaWt0b2slMjBjb250YSUyMGRpZ2l0YWwiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773898973),
('AYMNUyITwdtAHFGlLblv3cOAqrE7zPoEwBhmUukD', NULL, '198.244.183.95', 'Mozilla/5.0 (compatible; AhrefsBot/7.0; +http://ahrefs.com/robot/)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoidVpTRjJDSnRHeFRPMkFOWnV2NGRmODVQSnlWWVEwdEIwQTVycUg0RSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6Mjg6Imh0dHBzOi8vd3d3LmFtYW5pYS5pZC9qYWR3YWwiO3M6NToicm91dGUiO3M6MTg6InVzZXIuZXZlbnRzLmphZHdhbCI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773897098),
('BIKFPjRnhABDHDtp6bFry584pFW54Jy2Wr7hPtEm', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiRFhuT0lOWEU0ZzhYWDhyYWhOWEtLSVI5TUdIdjl0eUlqN0RjYmRpTyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzA6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9eGFuYXglMjB4ciI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773891187),
('bRBMGfWz7WjdFJcUdYTERJIkdK0YMh3Q7GicOUyn', NULL, '54.37.118.64', 'Mozilla/5.0 (compatible; AhrefsBot/7.0; +http://ahrefs.com/robot/)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiU2JkbzNGY1lqV2NsSmV1enR2d0JkUDZpa3IwMmZxUDlEMVVPSEI1MCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MjY6Imh0dHBzOi8vYW1hbmlhLmlkL2FydGljbGVzIjtzOjU6InJvdXRlIjtzOjE0OiJhcnRpY2xlcy5pbmRleCI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773902411),
('CdC2DOQI3vAnXfjcyYX2mDaRHZNRHD5WlL1HWN7l', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiNk0ySmFqTDhtRkhzQ3NGTUtGeEtFTW9VQWlVVG14WWhmZDQzTWlmSCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzI6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9d2F0ZXIlMjBwbGxzIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773894380),
('cl7oG5MjpQbQGy3wLrnenGZ5O9VVn6JLzuwbqdrx', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiYjBwOVhUdmcwYXplS3FZemNmWmZhTUVIS0NwaW5RZXkyU0d0UmlyRiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6Mjc6Imh0dHBzOi8vYW1hbmlhLmlkLmFwYWJhZS5pZCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773891127),
('cLJBB7Y7Ut9VUQW1QHOePcKoCjx5f78W44JPkFjf', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoidjZlSFhWMUl6SHAxSzNpRWJ4SHd4TTQ5QWhuWVM5R05OMVlwM0tSNyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTc6Imh0dHBzOi8vYW1hbmlhLmlkIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773896445),
('cY3oBW2Lhb3WwoTBITfQNFevp4mrZH4iQs3x5sjC', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiTzZTMFFqdHdyUlJIbEp2Rk1xZ1FyNjU2UzZYSWNnREcyOXl1VnJ1RiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDM6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9Z2VuZXJpYyUyMGNpYWxpcyUyMGNoYXQiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773902600),
('DfJErDclDBmNCnElL1vDBv9P4pNFbP4UdBAYZ9N6', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoic2lKbUxsNWo5b25YdVVDSlM3QU1Hcm82MllNQ1NCM01yZ20zT2FYYSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTc6Imh0dHBzOi8vYW1hbmlhLmlkIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773903647),
('eu8oDLEWiB3ti2cOOzf4B9d6fbipt4xisU44huYR', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiaFEwek50a2JPZlVKbk1JWWJGWWFJYU9PbW9VODJISVVTR1VnSmk5cSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDM6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9Y2hlYXBlc3QlMjBjaWFsaXMlMjBidXkiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773902597),
('eYMkEotN86cv7vd1QBYy1lpxzjnW0YgKgPXHrxY1', NULL, '51.68.236.92', 'Mozilla/5.0 (compatible; MJ12bot/v2.0.5; http://mj12bot.com/)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiZ3V2OG9jYmNNRWxlYzdRcEVrb3U3UTk4bWFMZUFGSkp3WlIyRlU2MiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTY6Imh0dHA6Ly9hbWFuaWEuaWQiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773895409),
('fa05Ve6jvl9g5MwpaHAnkbUZCM5bKNNp5Nsh8pk5', NULL, '189.192.47.205', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiTk81UEpwRE90M2htUWRRODBkUXkzSmt3M0dxMmg4akxzSDAyRm1DSiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDM6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9SlYlMkZFNzcwMzg1Lmh0bWwiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773895004),
('FCGhvClZbkkbfhSfUx0LveQhdy3hFyC8LUr0An7c', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiSGtldnhLaW1QdXN6NU0wdklRV2x0V2hySk1LSWx5WmpNOVlIbFFwQiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDY6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9Y2hlYXAlMjBwaGVudGVybWluZSUyMDM3LjUiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773900209),
('fLK0LmUFNVVD7qTjy17MH5eI4VBMW6q6T5bnSmoo', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoieVdjaTJ3TzNyaWNoeHB1OHJ1U0tmQnZvNEgxVDc5SWhpQzlyb0JpRiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NTg6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9bWV0Zm9ybWluJTIwb25saW5lJTIwbm8lMjBwcmVzY3JpcHRpb24iO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773894378),
('GJlYwnxNeiyzC4qPlp2cS1z94b5CkoyD3usQrEND', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiU1N4WVllZUpkT21rdElreFIwaHVnNmpUR05YSDVMbnBIdWdpdU1rZSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDU6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9YnV5JTIwbWVkcm94eXByb2dlc3Rlcm9uZSI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773900206),
('GP0GN9HsQ1oGbgIe5cIleLLLBhGy7qrTYSZQjzv5', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiVGhrTWtqNjdIVWJNR25GMWg2Y0lDdmgwdmdtemxMUXNPc2tCN0tNYSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDM6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9a2FtYWdyYSUyMDEwMG1nJTIwamVsbHkiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773897182),
('Gqenk2wjWbH56I2DrOZqMYUiN7LQQ2N0iIpHDIgt', NULL, '47.82.11.104', 'Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoid3R0R3BpdElyeWFoeUFCQk5Bd2tGWVd1TnZvSHRGRmRpbTRVMHBudSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9Y3IlMkZOMzI4MTA1My5odG1sIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773900051),
('HiTD91lN2gjk2dlGamuazj3FHRiwgstXgBjmv2s9', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiMThtUlFTemFIa3FhTVFuRjQwV0pLQ2tGb0M3anFlWklhYWdubXA1aSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NjY6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9ZWQlMjBtZWQlMjB0cmltaXglMjBmcm9tJTIwaW5kaWFuJTIwcGhhcm1hY2llcyI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773891182),
('hqBYPI9NAAG2ueLLzcYw9sn29V4h7M1gfp0laCiD', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoibGZSUnNveWs5Y3hHZHM1em5HVWpnT2RtZGNtZFMzMVVQMjR2ZVlIYSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6Mjc6Imh0dHBzOi8vYW1hbmlhLmlkLmFwYWJhZS5pZCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773900131),
('hRj5p7h8tyBbDhdcM7v7SoL1oUb43Dg0x6g6q7IP', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiRHhDeFRmU0M5QWw0NHEwN2lQNUxVV3NFTjE3TVdPcGlTcTNEMURuSSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NjM6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9MzYlMjBob3VyJTIwY2lhbGlzJTIwd2l0aG91dCUyMHByZXNjcmlwdGlvbiI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773891683),
('icu0u1yVNNMbbc6ZicijuZDHOLtaWUGZY6iHBOVW', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoibDZXMlRtc1Y0a3NFOFk4djRDeGRwZVpLcDBIWDhGajNjcHJNaTFXbiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9NSUyMG1nJTIwY2lhbGlzJTIwb25saW5lIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773900202),
('JPHuAq8j90tNZnxhKighkYkvDwnY9vZZOrXaqtdT', NULL, '144.31.25.220', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiMzhOZmVnYlNRNmhsQmhMUTYzV0FMMVJiYmtYOXZjN3hoYkhBQTNmQiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTc6Imh0dHBzOi8vYW1hbmlhLmlkIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773898288),
('Jy4zEXDidjhf4RDP3ynVEZidyehelSb4WGPs0Xiq', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiWTBqWGN0RkRNOFhSQ29sRWdCTkhEam9Zbk14M3RkY0RXOXYzbFZwdCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzI6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9dmFsaXVtJTIwMW1nIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773896697),
('KCbggAijuWGfGMaSLUQWhfRUDJ6stDbbM64Gxc5m', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiVDFLa0Q2RHVhdnBrYVZLUkxOblhvUERlZExPc2VqTXY5VTdvbHMwQyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6Mjc6Imh0dHBzOi8vYW1hbmlhLmlkLmFwYWJhZS5pZCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773896528),
('KjbrMFjpRybPY1LOKcZEvjKKfJ4gwhNlAwa6hQ9X', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiSnY2MlBad0N0VXZDN3BJWWFvSzFvcEZHcFVaSkJHTkZpTWhXdmVvWiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDY6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9b3JkZXIlMjBkb3hjeWNsaW5lJTIwMTAwbWciO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773897271),
('kKSduLTY9uQaVvOIbAA2N6FsYnjMIEOYWTfT7NH0', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiSnB0VUVxNVBJTUI1QlpkZFBzSHFGNDJYcFk4TFpPUHNvV1VQZnpKMCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NTk6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9b24lMjBsaW5lJTIwcGhhcm1hY3klMjBsZXZpdHJhJTIwY2FuYWRhIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773891185),
('kme0UqHrZ5BR3Yx2IRdTBDzzI63USHlyEYNJPSMD', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiT1UwblhaMm8zZWdydExBSjdSa0dOSWNYV0psRzlrTVFuZEFON0dwRyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NTM6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9ZG9lcyUyMGFldG5hJTIwcGF5JTIwZm9yJTIwY2lhbGlzIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773901790),
('KNUfIaypr4ZiyhiefLXuXOvwpzdTmIA20kaIYpYQ', NULL, '15.204.182.106', 'Mozilla/5.0 (X11; CrOS x86_64 14541.0.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiMnZZOXJudkNoNUhnNjVCUlYzd1dpdzV2U1hCNnU5eUU2VTVxbFpoeSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MjE6Imh0dHBzOi8vd3d3LmFtYW5pYS5pZCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773891696),
('L6L3nVmHWVinb16gZdXCdulZ5j7fc2c3HywTN8Tv', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoidEZNaXYyWHVORW1kZ3EwaTV3U2o3UUhsbkZCb2lzbjhETHpqUDBaayI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NjI6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9YmVzdCUyMG9ubGluZSUyMHZpYWdyYSUyMG5vJTIwcHJlc2NyaXB0aW9uIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773891685),
('LfCfC1H9xYBwKhyIKeY482PdJGME8Ykg7mcWeTKV', NULL, '14.173.0.25', 'Mozilla/5.0 (X11; U; Linux i686; en-US) AppleWebKit/532.2 (KHTML, like Gecko) Chrome/4.0.221.0 Safari/532.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiaG1HcHp1ZFY4RWRPN05xdGUyVmxjYXdqcVRyUmM1UkpvaWJ3eFJFdiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MjU6Imh0dHBzOi8vYW1hbmlhLmlkLz9jYXQ9MzIiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773901010),
('lOYZStL4sCwa84OaIlt3vE8UbBVLaoqEqOGomlKQ', NULL, '46.36.123.85', 'fasthttp', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiU0VOa2ZSMnZCV01hZk15UTRHMWpza1ZyTjhEWEt6bEZPbDBEeENvRiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTY6Imh0dHA6Ly9hbWFuaWEuaWQiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773898704),
('Mf8ZXVCdgOdZFXirp7o4Aox0AXEDBbSfG0JNkc3q', NULL, '129.222.147.30', 'Mozilla/5.0 (X11; CrOS x86_64 14816.131.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoidHRqdWZTdHpQTXV2eWFNT3lZb0gzTjN5clp5Q2oyZGZpM1RaNTdjcCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9ZG4lMkZOODcxOTc1OS5odG1sIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773903695),
('mFu9te94TIx7CORtelzkEaUtAAmMYLVeSFg1tM0d', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiMlZUYlZLRnFlMjQyaVR5UDMyZGRxUUFteUUwckt0ajJNcXFSdVZPYSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9YnV5JTIwem92aXJheCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773899212),
('MmmHqV1AxcjgLKnk54pYDnna1uHoE3uHtC3xXTwq', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiOXJlazJncDc2QjJYZnRlMFQxbkxCYVg1RXk1T3oyNXI3RnNxWTd3SiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzI6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9dmlhZ3JhJTIwZmFxIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773901918),
('n0BxiDiqR2RYydfYh4riYrLbyQUc4kDhRIisj9J0', NULL, '198.244.168.49', 'Mozilla/5.0 (compatible; AhrefsBot/7.0; +http://ahrefs.com/robot/)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiVjlCQk9OZ1NIbndtc0hBNVFENTA4cEJHa3ozd1pzZHVyaVNOZmpqQyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MjE6Imh0dHBzOi8vd3d3LmFtYW5pYS5pZCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773900585),
('nFbMKyGp5XtDzmMvrKBslDU9TrsjePJxXcLEFybk', NULL, '45.162.138.185', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiaUtXS3lkRlZJQ0dlcWlKUHhaaTJnYkk0dlpCSmNRZTVXMkVtRHgyYiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9Ym8lMkZIMzYxMzA4OC5odG1sIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773896785),
('nr2v4hKVituAbqUwOueaHtBjmI40ffUtcXpxnC1D', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiTnlFb2RrdVZYQ04yQWVRV3J4QUM2Z29vNDZHbDI3TDFZQWtMQlpXaiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTc6Imh0dHBzOi8vYW1hbmlhLmlkIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773898250),
('nyah31QwQ1lVVp18HVqwDOZqYFsQFHu73thR6Y9X', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiNTNObjRjZXNWNVM3Z3ZvNDNTZzJqQmpzNkN4NUoyNlN3UzVzMUp5YyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzU6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9cHJlenppJTIwdmlhZ3JhIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773891687),
('o2Be6MjrrxuD4RAJnvCwm33N07P2gUfDKAiJv63x', NULL, '36.255.34.14', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiRmtFMUhZdW9aeEFsUXZFM29TREljeDBIbUlNVXhiT2k5cnVNUUVxeCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDM6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9S0slMkZMNzQwMTMyLmh0bWwiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773900530),
('omUhxWj6d8RLKZ2G8N0kX5rEiHg3bg3VtLups7sb', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiQmFQN3hhV0d4SXJ2Q3VnOUtxZXRyd3haYzhEdzZzTnN5WlVIam1tdyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTc6Imh0dHBzOi8vYW1hbmlhLmlkIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773900049),
('PFNiTofdAkdpdioIO5U9X4TPjW80aehYu4OKKJmx', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiV3BDWlJqb21GdzdKYVBSNmJqWmNUYVJvU2s4Q1h4SlNDZ3JWUGZEdCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDM6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9dHJhbWFkb2wlMjBwb3RlbnRpYXRvcnMiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773891860),
('PiKKg0nc8IKiBpRj6GEVyS6OD0jyQwFE6O4KPHeu', NULL, '189.49.77.20', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3.1 Safari/605.1.15', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiR0FwMUcwdkJIS3BXU1d6bEtQTTFYbjZ0bVlzUXJkVzk0d3FOT2lqaSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9ZHMlMkZXNTk4OTI5My5odG1sIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773899940),
('PXWJfR6rWsX2uX2uGK0DnHp1S35MchPP7ctvAlCq', NULL, '189.91.177.5', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Safari/605.1.15', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiczdIaUsxenczOElYckswcDZxZ2xBZ2Z1MmRBQUpqMldtRXAyRlJ0byI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9Y3glMkZMMjU5NzU3Mi5odG1sIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773893084),
('rx8HlemLoKMopJYG2Nl3lWaDsPAyQvIiwAMyWPPH', NULL, '176.97.63.89', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiam5PcDBCeDV5VW9uMnpUMWlPaXRkejlzR2tSRnc0T2wxcENlemYzeSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9cHMlMkZZNjE4MDQyMy5odG1sIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773896904),
('sSJQa46Yghj0fcYmHf2L6QMt5UiF0MMNELOWJp9o', NULL, '51.195.183.194', 'Mozilla/5.0 (compatible; AhrefsBot/7.0; +http://ahrefs.com/robot/)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiYUhTSktpRDdOMzdHSjhOc2o3TkpTOEZFRzlIWDZrYlZVaEMwZzZ5QyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTY6Imh0dHA6Ly9hbWFuaWEuaWQiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773893720),
('SSWSNOlwd1lQHF3Tu81ZaQ5XEKuM2GGmV5avW4O2', NULL, '47.82.11.76', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:137.0) Gecko/20100101 Firefox/137.0', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiR0cxTE1laDlFNXVvclFmaVZmYlo0eFBUTTZ1TkFSNXAzY1VrdVJlQiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9a2ElMkZTNjExMDA5OS5odG1sIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773899864),
('sWBlbk9nyxBOFfuQVrWhdrzfCBUmK1R5NWUtQhlq', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiVk9LZXhTZjBLVzNaR292SGp3eUowbjdscTVuZkZxaUZFZ2pXcWFYNiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDM6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9YnV5JTIwdmlhZ3JhJTIwMTUwJTIwbWciO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773902595),
('sxIjcWO7VVEKVoAnELXMXVdPuczVROqNZjFhA3Sx', NULL, '168.227.20.175', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiZkdUelNoQlNNS2c0WjQ1OXdZSVEyMnQ3aWRidUdNSXVpZzZQVEJtMSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9aGclMkZDNDI3MjcyNC5odG1sIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773895795),
('T7kpUNr1wc4ymoZ4sYEK6xowKZZm3SHgh3kjbaup', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiT2ZnSG1RTXQ5NXd1S0ZnUVdUeWxqMTMzYThCRTZMeUlva1dLekdJcyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzA6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9Z29sZCUyMGZseSI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773901793),
('ToC9CLjucevrYHKlLVDyrlqurWb0714IbnWaTcWQ', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiaU9sdGJOWllrcTkzR3lKa2dpeWwxS3NCNmYyNUFLNUdDTmFwWThrdCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzI6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9dmlhZ3JhJTIwZ2VsIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773899217),
('TrpReoDVFgbFTFyCsqJhf4Ly7iFQAsRxBOxrDUNz', NULL, '179.96.166.53', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiTzh3bm5pSWRaOWhYbmtBUndBTmlKTzU3bWd4djhRcGV1eGNPMUZRciI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDM6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9QkolMkZGNzA4ODEwLmh0bWwiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773893257),
('TZnWSYttoZ83Gx2tCH06wZs8VW1JDbfyZHwPpS1o', NULL, '51.68.111.203', 'Mozilla/5.0 (compatible; MJ12bot/v2.0.5; http://mj12bot.com/)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiZ3JSa01Vc2J3Q25Sa2w5eG1RYXM3Sk1xNWxqbHgxOXBCaUU1T3dDVCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MjA6Imh0dHA6Ly93d3cuYW1hbmlhLmlkIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773890392),
('uEDZgPLyVvzPV0X3gWNzEYOTQQl7FE4XTsjTqtuV', NULL, '81.131.119.177', 'Mozilla/5.0 (Windows NT 6.1; WOW64; rv:23.0) Gecko/20130406 Firefox/23.0', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiaE9TZnVMWXFzSHVDVDhpY1dKMXp4NE5ZalFvZWlFQ2Q0VHJvbmViWSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MjM6Imh0dHA6Ly9hbWFuaWEuaWQvP2NhdD0zIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773891623),
('UjMkRpU7kJOLKc3OHoeJY3B0oSXsKjtlpvJ36tft', NULL, '47.82.11.159', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoielRVUGFaUkNCYW04NDFuUWM1NjJpN3JZV2RsenhQeUM0d0RTVXhOVCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDM6Imh0dHBzOi8vYW1hbmlhLmlkLz90YXJnZXQ9R1AlMkZaNzgzNjk0Lmh0bWwiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773893421),
('USa9N9Jx9JnsOGG0IPPWOOhHFg0w1SnLy2vImvRm', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiYkx6TG9KY3V0WlBtanJmN3lsOHlQaWVJOTVKRzBzWjM2bGt5OE1vRiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDY6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9bGV2aXRyYSUyMHZpYWdyYSUyMGdlbmVyaWMiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773900212),
('uW1F63FJsppSw6A989RZdfCrH9lxeeJvrD1rPZNA', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoicmFWUWYxYTlIcm10QTdCUElFWHZMaUlQS1NJRmNCV3hwTmdWTHA0diI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTc6Imh0dHBzOi8vYW1hbmlhLmlkIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773892846),
('VC3g0GdtTU6r6NfvH7LxeD1ouA0mSjGJoLdpsHnb', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoib2hVRzNkcG1mRmlEWEp0RE5uNG9lcjBrRlZQS2hYSXFIeGtBOWpQaiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9bmV4dCUyMGRheSUyMHZpYWdyYSUyMHVrIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773900215),
('W6wuaihPC3rSxxgLWiKeOdyu5ypQXxyQIYv4bF6C', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiakhWOXZwNENEaDE1S2doa1NrVm9hcGNVZnkwckJTbnBia1JXRkZzMyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDY6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9c3Vic3RpdHV0ZXMlMjBmb3IlMjBjaWFsaXMiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773900219),
('X12O6ADt0eANhZ9s9fVtDnZNqaahrYPg26oUT8fL', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoib1ZsR0NLdlhIN0FuUWZDNWtGQUxqWFFIdkFlY2g2MDRXVGd2ckUydCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDA6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9cGhlbnRlcm1pbmUlMjBmYXN0aW4iO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773894098),
('x1OHfYWYk7ZpLKQDlst3oMXRTseyQ0elKvF9nQOq', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiNmVPbDdleHNqa1RjQ29XYWZkNVFwVzM1bTQ3ajR0ZlkxZUVLZkJGSyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDY6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9Y2hlYXAlMjBhY3ljbG92aXIlMjBvbmxpbmUiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773897268),
('Xbo6HqEU7BDCUavgKNWRgbBJtDKTXhutg27RVPT1', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiMmhaY1V2Z3VBaTR5bklXeWduclpPT3dweEd4M212ekdJMEdUTGg3SSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDk6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9Y2hlYXAlMjBrYW1hZ3JhJTIwcGF5cGFsJTIwdWsiO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773892027),
('XgWxQRrFZWWTYXJ1r8XckAaoH48XDCBZgGd2xoQe', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiVGt5bEJMWWZhN2dOaE45bUN4YkQxQ2FaMjkzbXR5dDFTdk9lUGxTYSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTc6Imh0dHBzOi8vYW1hbmlhLmlkIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773891045),
('YDe5Xcr2AyOmpefojv0NMEhWMy3CO6fhMEz5KctQ', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiYjlwQlloTVdnTWNDdXN1QjFiMVI5aUJrVFNFUlJHZjBLYkdUcW43NCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MTc6Imh0dHBzOi8vYW1hbmlhLmlkIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773901850),
('yEvk0yq5sy68HAchE3OakR8HzUP3lTgDiaj8bl84', NULL, '85.208.96.197', 'Mozilla/5.0 (compatible; SemrushBot/7~bl; +http://www.semrush.com/bot.html)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiWmwyTzNhRU9jZ2ZNOW5Ybjh5TTExbksweE5keGdtSTA4TTI5cDJXdiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NzY6Imh0dHBzOi8vYW1hbmlhLmlkLz9zPSVFMiU5RSU5QiUyMGFjYzYudG9wJTIwZXJoYWx0ZW4lMjBlaW5lcyUyMEF6dXJlJTIwa29udG8iO3M6NToicm91dGUiO3M6NDoiaG9tZSI7fXM6NjoiX2ZsYXNoIjthOjI6e3M6Mzoib2xkIjthOjA6e31zOjM6Im5ldyI7YTowOnt9fX0=', 1773897269),
('YHgYr9SKAJgHpBmwlV3vqe4PM1bI7m2Vx0ow6cV5', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiOUNaZUFWbFoyakt3eEJGdHhEaUZOdUJaVHV6TlBKSWNNUmw0Rkh2SSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzI6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9YnV5JTIwY2lhbGFzIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773899209),
('YsHzLCUDvICU10Qw6Mt3bz8Y4SdgKy2G2GhHiJfv', NULL, '103.79.244.233', 'Python/3.11 aiohttp/3.9.2', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiYU1EZlB5bDdKaTlDYVcxak1oNWFvNjVPOFlmYXl6S1I0R0tsVGlRNyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6Mjc6Imh0dHBzOi8vYW1hbmlhLmlkLmFwYWJhZS5pZCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==', 1773903731),
('zC2DZLbBsWuN5ZzvbngCwS0KNbrrUMiCtQ4nPlSa', NULL, '185.191.171.13', 'Mozilla/5.0 (compatible; SemrushBot/7~bl; +http://www.semrush.com/bot.html)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoiMFU0UE9CSWhGeWFBMkJ4SHVXdHZLUlU5OHFQMTVhSHBWQmkzYlBIbCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MjI2OiJodHRwOi8vYW1hbmlhLmlkLz9zPSVEMCU5QyVEMCVCNSVEMCVCRCVEMSU4RiUyQiVEMCVCNyVEMCVCRSVEMCVCMiVEMSU4MyVEMSU4MiUyQiVEMCU5QSVEMSU4NSVEMCVCMCVEMCVCRCUyQiVEMSU4MSVEMCVCQyVEMCVCRSVEMSU4MiVEMSU4MCVEMCVCNSVEMSU4MiVEMSU4QyUyQiVEMCVCRSVEMCVCRCVEMCVCQiVEMCVCMCVEMCVCOSVEMCVCRCUyQnNtb3RyZXRvbmxheW5maWxteWlzZXJpYWx5LnJ1IjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773902275),
('ZMoKt8ahOGMerOiS03xXGZV1BasCLyetFLumQJVV', NULL, '216.244.66.248', 'Mozilla/5.0 (compatible; DotBot/1.2; +https://opensiteexplorer.org/dotbot; help@moz.com)', 'YTozOntzOjY6Il90b2tlbiI7czo0MDoicXlTNW42bVpNS0M2TUIyRVFwMHlxRURINUFnS2VvcmtNMU5ER0NpVCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6NDQ6Imh0dHA6Ly9hbWFuaWEuaWQvP3E9YnV5JTIwcmV0aW4lMjBhJTIwb25saW5lIjtzOjU6InJvdXRlIjtzOjQ6ImhvbWUiO31zOjY6Il9mbGFzaCI7YToyOntzOjM6Im9sZCI7YTowOnt9czozOiJuZXciO2E6MDp7fX19', 1773891923);

-- --------------------------------------------------------

--
-- Table structure for table `speakers`
--

CREATE TABLE `speakers` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `event_id` bigint(20) UNSIGNED NOT NULL,
  `name` varchar(255) NOT NULL,
  `role` varchar(255) DEFAULT NULL,
  `photo` varchar(255) NOT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `name` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `email_verified_at` timestamp NULL DEFAULT NULL,
  `password` varchar(255) NOT NULL,
  `role` enum('admin','user') NOT NULL DEFAULT 'user',
  `remember_token` varchar(100) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `users`
--

INSERT INTO `users` (`id`, `name`, `email`, `email_verified_at`, `password`, `role`, `remember_token`, `created_at`, `updated_at`) VALUES
(1, 'Administrator EMS', 'admin@edutech.com', NULL, '$2y$12$pqJELsl9bDgMScfX9EzaQ.Oqc4227KrpVmrRalR/Ao1WLrNXnYDpa', 'admin', NULL, '2026-02-18 21:59:04', '2026-02-18 21:59:04'),
(14, 'Nurul Hidayat', 'nurul@unsoed.ac.id', NULL, '$2y$12$hCXAYcqJjuAEBIwScOh3KewiMbQkiqN6uT/zgDt9FrjtqIz1oPiFy', 'user', NULL, '2026-03-05 11:38:04', '2026-03-05 11:38:04');

--
-- Indexes for dumped tables
--

--
-- Indexes for table `articles`
--
ALTER TABLE `articles`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `articles_slug_unique` (`slug`);

--
-- Indexes for table `bank_accounts`
--
ALTER TABLE `bank_accounts`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `bank_account_event`
--
ALTER TABLE `bank_account_event`
  ADD PRIMARY KEY (`id`),
  ADD KEY `bank_account_event_event_id_foreign` (`event_id`),
  ADD KEY `bank_account_event_bank_account_id_foreign` (`bank_account_id`);

--
-- Indexes for table `cache`
--
ALTER TABLE `cache`
  ADD PRIMARY KEY (`key`);

--
-- Indexes for table `cache_locks`
--
ALTER TABLE `cache_locks`
  ADD PRIMARY KEY (`key`);

--
-- Indexes for table `events`
--
ALTER TABLE `events`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `failed_jobs`
--
ALTER TABLE `failed_jobs`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `failed_jobs_uuid_unique` (`uuid`);

--
-- Indexes for table `jobs`
--
ALTER TABLE `jobs`
  ADD PRIMARY KEY (`id`),
  ADD KEY `jobs_queue_index` (`queue`);

--
-- Indexes for table `job_batches`
--
ALTER TABLE `job_batches`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `materials`
--
ALTER TABLE `materials`
  ADD PRIMARY KEY (`id`),
  ADD KEY `materials_event_id_foreign` (`event_id`);

--
-- Indexes for table `migrations`
--
ALTER TABLE `migrations`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `password_reset_tokens`
--
ALTER TABLE `password_reset_tokens`
  ADD PRIMARY KEY (`email`);

--
-- Indexes for table `registrations`
--
ALTER TABLE `registrations`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `registrations_ticket_code_unique` (`ticket_code`),
  ADD KEY `registrations_event_id_foreign` (`event_id`),
  ADD KEY `registrations_user_id_foreign` (`user_id`);

--
-- Indexes for table `sessions`
--
ALTER TABLE `sessions`
  ADD PRIMARY KEY (`id`),
  ADD KEY `sessions_user_id_index` (`user_id`),
  ADD KEY `sessions_last_activity_index` (`last_activity`);

--
-- Indexes for table `speakers`
--
ALTER TABLE `speakers`
  ADD PRIMARY KEY (`id`),
  ADD KEY `speakers_event_id_foreign` (`event_id`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `users_email_unique` (`email`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `articles`
--
ALTER TABLE `articles`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT for table `bank_accounts`
--
ALTER TABLE `bank_accounts`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT for table `bank_account_event`
--
ALTER TABLE `bank_account_event`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT for table `events`
--
ALTER TABLE `events`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=7;

--
-- AUTO_INCREMENT for table `failed_jobs`
--
ALTER TABLE `failed_jobs`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `jobs`
--
ALTER TABLE `jobs`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `materials`
--
ALTER TABLE `materials`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT for table `migrations`
--
ALTER TABLE `migrations`
  MODIFY `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=17;

--
-- AUTO_INCREMENT for table `registrations`
--
ALTER TABLE `registrations`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT for table `speakers`
--
ALTER TABLE `speakers`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=15;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `bank_account_event`
--
ALTER TABLE `bank_account_event`
  ADD CONSTRAINT `bank_account_event_bank_account_id_foreign` FOREIGN KEY (`bank_account_id`) REFERENCES `bank_accounts` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `bank_account_event_event_id_foreign` FOREIGN KEY (`event_id`) REFERENCES `events` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `materials`
--
ALTER TABLE `materials`
  ADD CONSTRAINT `materials_event_id_foreign` FOREIGN KEY (`event_id`) REFERENCES `events` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `registrations`
--
ALTER TABLE `registrations`
  ADD CONSTRAINT `registrations_event_id_foreign` FOREIGN KEY (`event_id`) REFERENCES `events` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `registrations_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `speakers`
--
ALTER TABLE `speakers`
  ADD CONSTRAINT `speakers_event_id_foreign` FOREIGN KEY (`event_id`) REFERENCES `events` (`id`) ON DELETE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
