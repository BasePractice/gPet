CREATE TABLE classes
(
    id         SERIAL    NOT NULL,
    name       VARCHAR   NOT NULL,
    table_name VARCHAR   NOT NULL,                                                                  -- Values table name
    current    INTEGER                                                                   DEFAULT 0, -- Current version
    status     VARCHAR   NOT NULL CHECK ( status IN ('DRAFT', 'PUBLISHED', 'ARCHIVED') ) DEFAULT 'DRAFT',
    title      VARCHAR                                                                   DEFAULT NULL,
    updated_at TIMESTAMP NOT NULL                                                        DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL                                                        DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (name)
);
INSERT INTO classes(name, table_name, title)
VALUES ('sex', 'class_sex', 'Пол человека');
CREATE TABLE class_sex
(
    id         SERIAL    NOT NULL,
    key        VARCHAR   NOT NULL,
    value      VARCHAR   NOT NULL,
    version    INTEGER   NOT NULL                                                    DEFAULT 0,
    status     VARCHAR   NOT NULL CHECK ( status IN ('DRAFT', 'PUBLISHED', 'SKIP') ) DEFAULT 'DRAFT',
    before_at  TIMESTAMP                                                             DEFAULT NULL,
    after_at   TIMESTAMP                                                             DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL                                                    DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL                                                    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (key, value, version)
);
CREATE TABLE class_values_changes
(
    id         SERIAL    NOT NULL,
    class      VARCHAR   NOT NULL,
    class_id   INTEGER   NOT NULL,
    version    INTEGER   NOT NULL,
    key        VARCHAR            DEFAULT NULL,
    changes    VARCHAR            DEFAULT NULL,
    action     VARCHAR   NOT NULL CHECK ( action IN ('CREATE', 'STATUS', 'AFTER') ),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION fn_class_sex_after_insert() RETURNS TRIGGER AS
$$
BEGIN
    INSERT INTO class_values_changes(class, class_id, version, key, changes, action)
    VALUES ('class_sex', NEW.id, NEW.version, NEW.key, NEW.value, 'CREATE');
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';


CREATE TRIGGER class_sex_after_insert
    AFTER INSERT
    ON class_sex
    FOR EACH ROW
EXECUTE FUNCTION fn_class_sex_after_insert();

CREATE OR REPLACE FUNCTION fn_class_sex_after_update_status() RETURNS TRIGGER AS
$$
BEGIN
    INSERT INTO class_values_changes(class, class_id, version, key, action, changes)
    VALUES ('class_sex', NEW.id, NEW.version, NEW.key, 'STATUS', concat(OLD.status, '->', NEW.status));
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

END;
CREATE TRIGGER class_sex_after_update_status
    AFTER UPDATE
    ON class_sex
    FOR EACH ROW
    WHEN (NEW.status != OLD.status)
EXECUTE FUNCTION fn_class_sex_after_update_status();

CREATE OR REPLACE FUNCTION fn_class_sex_after_update_after() RETURNS TRIGGER AS
$$
BEGIN
    INSERT INTO class_values_changes(class, class_id, version, key, action)
    VALUES ('class_sex', NEW.id, NEW.version, NEW.key, 'AFTER');
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER class_sex_after_update_after
    AFTER UPDATE
    ON class_sex
    FOR EACH ROW
    WHEN (NEW.after_at != OLD.after_at)
EXECUTE FUNCTION fn_class_sex_after_update_after();

INSERT INTO class_sex(key, value)
VALUES ('m', 'мужской');
INSERT INTO class_sex(key, value)
VALUES ('f', 'женский');
INSERT INTO class_sex(key, value)
VALUES ('n', 'не определен');
