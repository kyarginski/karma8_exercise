CREATE TABLE IF NOT EXISTS cache (
checksum VARCHAR(64) NOT NULL UNIQUE,
filename VARCHAR(255) NOT NULL,
expired_at TIMESTAMP   NOT NULL DEFAULT timezone('utc'::text, now() + '1 day'::interval)
);

COMMENT ON TABLE cache IS 'Table for storing cache of files';
COMMENT ON COLUMN cache.checksum IS 'Unique hash of the file as a string';
COMMENT ON COLUMN cache.filename IS 'Name of the file';
COMMENT ON COLUMN cache.expired_at IS 'Date and time of the record expiration';
