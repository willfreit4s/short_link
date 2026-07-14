BEGIN;

CREATE TABLE IF NOT EXISTS links (
    id varchar(12) NOT NULL PRIMARY KEY,
    original_url text NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
)

COMMIT;