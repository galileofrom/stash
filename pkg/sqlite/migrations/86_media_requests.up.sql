-- Galileo fork: media request feature.
-- Migration number 90000 chosen to leave headroom for upstream stash migrations
-- without rename conflicts on merge.

CREATE TABLE `media_requests` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `title` varchar(510) NOT NULL,
  `studio` varchar(510),
  `external_id` varchar(255),
  `external_source` varchar(64),
  `metadata_json` text,
  `status` varchar(32) NOT NULL DEFAULT 'pending',
  `requested_by` varchar(255),
  `notes` text,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL
);

CREATE INDEX `idx_media_requests_status` ON `media_requests` (`status`);
CREATE INDEX `idx_media_requests_external` ON `media_requests` (`external_source`, `external_id`);

CREATE TABLE `media_request_releases` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `request_id` integer NOT NULL,
  `indexer` varchar(255) NOT NULL,
  `indexer_id` integer,
  `guid` varchar(510) NOT NULL,
  `title` varchar(1020) NOT NULL,
  `download_url` text,
  `magnet_url` text,
  `info_url` text,
  `size` bigint,
  `seeders` integer,
  `leechers` integer,
  `publish_date` datetime,
  `category` varchar(255),
  `protocol` varchar(32),
  `raw_json` text,
  `selected` boolean NOT NULL DEFAULT 0,
  `grab_status` varchar(32),
  `grab_error` text,
  `created_at` datetime NOT NULL,
  FOREIGN KEY (`request_id`) REFERENCES `media_requests`(`id`) ON DELETE CASCADE
);

CREATE INDEX `idx_media_request_releases_request_id` ON `media_request_releases` (`request_id`);
CREATE UNIQUE INDEX `idx_media_request_releases_guid` ON `media_request_releases` (`request_id`, `guid`);

CREATE TABLE `media_request_downloads` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `request_id` integer NOT NULL,
  `release_id` integer NOT NULL,
  `download_client` varchar(64),
  `download_id` varchar(255),
  `status` varchar(32) NOT NULL DEFAULT 'queued',
  `progress` real NOT NULL DEFAULT 0,
  `eta_seconds` integer,
  `output_path` text,
  `error` text,
  `started_at` datetime,
  `completed_at` datetime,
  `imported_at` datetime,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  FOREIGN KEY (`request_id`) REFERENCES `media_requests`(`id`) ON DELETE CASCADE,
  FOREIGN KEY (`release_id`) REFERENCES `media_request_releases`(`id`) ON DELETE CASCADE
);

CREATE INDEX `idx_media_request_downloads_status` ON `media_request_downloads` (`status`);
CREATE INDEX `idx_media_request_downloads_request_id` ON `media_request_downloads` (`request_id`);
