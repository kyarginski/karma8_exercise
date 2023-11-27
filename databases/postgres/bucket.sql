CREATE TABLE IF NOT EXISTS bucket (
	id bigint not null,
	address TEXT not null,
	active_sign bool default true
);

ALTER TABLE bucket
    ADD CONSTRAINT bucket_key PRIMARY KEY (id);

COMMENT ON TABLE bucket IS 'A table for store buckets info. Author: Viktor Kyarginskiy';
COMMENT ON COLUMN bucket.id IS 'ID of the bucket';
COMMENT ON COLUMN bucket.address IS 'Address of the bucket';
COMMENT ON COLUMN bucket.active_sign IS 'Is bucket active?';

INSERT INTO bucket (id, address, active_sign) VALUES
(1, 'http://localhost:8261', true),
(2, 'http://localhost:8262', true),
(3, 'http://localhost:8263', true);

