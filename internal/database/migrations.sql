CREATE TABLE IF NOT EXISTS schema_migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    applied_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_name TEXT NOT NULL,
    android_latest_version TEXT NOT NULL DEFAULT '1.0.0',
    android_min_version TEXT NOT NULL DEFAULT '1.0.0',
    android_force_update INTEGER NOT NULL DEFAULT 0,
    ios_latest_version TEXT NOT NULL DEFAULT '1.0.0',
    ios_min_version TEXT NOT NULL DEFAULT '1.0.0',
    ios_force_update INTEGER NOT NULL DEFAULT 0,
    maintenance_mode INTEGER NOT NULL DEFAULT 0,
    maintenance_message TEXT NOT NULL DEFAULT '',
    banner_enabled INTEGER NOT NULL DEFAULT 0,
    banner_message TEXT NOT NULL DEFAULT '',
    api_url TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS feature_flags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE,
    enabled INTEGER NOT NULL DEFAULT 0
);
