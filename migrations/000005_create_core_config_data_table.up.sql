CREATE TABLE IF NOT EXISTS core_config_data (
    id VARCHAR(36) PRIMARY KEY,
    scope VARCHAR(20) NOT NULL DEFAULT 'default',
    scope_id VARCHAR(36) NOT NULL DEFAULT '0',
    path VARCHAR(255) NOT NULL,
    value TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    CONSTRAINT uq_scope_scope_id_path UNIQUE (scope, scope_id, path)
);

CREATE INDEX IF NOT EXISTS idx_config_path ON core_config_data(path);

-- Seed default initial configurations (Magento-style)
INSERT INTO core_config_data (id, scope, scope_id, path, value, created_at, updated_at)
VALUES 
    ('cfg-base-url', 'default', '0', 'web/unsecure/base_url', 'http://localhost:3000', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cfg-secure-base-url', 'default', '0', 'web/secure/base_url', 'http://localhost:3000', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cfg-api-base-url', 'default', '0', 'web/api_base_url', 'http://localhost:8080/api/v1', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cfg-store-name', 'default', '0', 'general/store_information/name', 'CodebaseGo Store', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cfg-timezone', 'default', '0', 'general/locale/timezone', 'Asia/Ho_Chi_Minh', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (scope, scope_id, path) DO NOTHING;
