CREATE TABLE IF NOT EXISTS metadata (
    uuid UUID PRIMARY KEY,
    checksum VARCHAR(64) NOT NULL UNIQUE,
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(255) NOT NULL,
    bucket_ids BIGINT[],
    created_at TIMESTAMP   NOT NULL DEFAULT timezone('utc'::text, now())
);

COMMENT ON TABLE metadata IS 'Table for storing metadata of files';
COMMENT ON COLUMN metadata.uuid IS 'Unique identifier of the file in UUID format';
COMMENT ON COLUMN metadata.checksum IS 'Unique hash of the file as a string';
COMMENT ON COLUMN metadata.filename IS 'Name of the file';
COMMENT ON COLUMN metadata.content_type IS 'Content Type of the file';
COMMENT ON COLUMN metadata.bucket_ids IS 'Array of bucket ids where the file is stored';
COMMENT ON COLUMN metadata.created_at IS 'Date and time of the record creation';
