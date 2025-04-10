CREATE TABLE tokens
(
    id    UUID         NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    hash  VARCHAR(128) NOT NULL,
    data  BYTEA        NOT NULL,
    title VARCHAR      NOT NULL,
    UNIQUE (hash)
);
CREATE TABLE keys
(
    id       UUID         NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    hash     VARCHAR(128) NOT NULL,
    num      BIGINT       NOT NULL,
    token_id UUID         NOT NULL,
    user_id  UUID         NOT NULL,
    UNIQUE (num, token_id, user_id),
    UNIQUE (hash)
);
