--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

-- initialize the database

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

CREATE TABLE IF NOT EXISTS cache (
                                     checksum VARCHAR(64) NOT NULL UNIQUE,
                                     filename VARCHAR(255) NOT NULL,
                                     expired_at TIMESTAMP   NOT NULL DEFAULT timezone('utc'::text, now() + '1 day'::interval)
);

COMMENT ON TABLE cache IS 'Table for storing cache of files';
COMMENT ON COLUMN cache.checksum IS 'Unique hash of the file as a string';
COMMENT ON COLUMN cache.filename IS 'Name of the file';
COMMENT ON COLUMN cache.expired_at IS 'Date and time of the record expiration';

CREATE TABLE IF NOT EXISTS metadata (
                                        uuid UUID PRIMARY KEY,
                                        checksum VARCHAR(64) NOT NULL UNIQUE,
                                        filename VARCHAR(255) NOT NULL,
                                        created_at TIMESTAMP   NOT NULL DEFAULT timezone('utc'::text, now())
);

COMMENT ON TABLE metadata IS 'Table for storing metadata of files';
COMMENT ON COLUMN metadata.uuid IS 'Unique identifier of the file in UUID format';
COMMENT ON COLUMN metadata.checksum IS 'Unique hash of the file as a string';
COMMENT ON COLUMN metadata.filename IS 'Name of the file';
COMMENT ON COLUMN metadata.created_at IS 'Date and time of the record creation';
