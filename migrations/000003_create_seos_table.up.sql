CREATE TABLE IF NOT EXISTS seos (
    id VARCHAR(36) PRIMARY KEY,
    slug VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    keywords VARCHAR(500),
    canonical_url VARCHAR(500),
    og_title VARCHAR(255),
    og_description TEXT,
    og_image VARCHAR(500),
    og_type VARCHAR(50) DEFAULT 'website',
    twitter_card VARCHAR(50) DEFAULT 'summary_large_image',
    robots VARCHAR(100) DEFAULT 'index, follow',
    structured_data TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_seos_slug ON seos(slug);
CREATE INDEX IF NOT EXISTS idx_seos_created_at ON seos(created_at);
