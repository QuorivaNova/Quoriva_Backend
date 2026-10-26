BEGIN;
CREATE TABLE foundation_metadata (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    version integer NOT NULL CHECK (version = 1),
    created_at timestamptz NOT NULL DEFAULT now()
);
INSERT INTO foundation_metadata (singleton, version) VALUES (true, 1);
COMMIT;
