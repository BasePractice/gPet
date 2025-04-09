CREATE TABLE token
(
    id    UUID         NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    hash  VARCHAR(128) NOT NULL,
    data  BYTEA        NOT NULL,
    title VARCHAR      NOT NULL,
    UNIQUE (hash)
)
