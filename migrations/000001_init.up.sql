CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    token VARCHAR(10) NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_urls_token ON urls(token);
CREATE UNIQUE INDEX idx_urls_original_url ON urls(original_url);