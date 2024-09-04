BEGIN;

DROP TRIGGER IF EXISTS urls_update_updated_at ON urls;

DROP FUNCTION IF EXISTS update_timestamp();

DROP TABLE IF EXISTS urls;

END;