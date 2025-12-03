CREATE TABLE index_test (
    flag BOOLEAN,
    value1 TEXT,
    value2 TEXT,
    value3 TEXT
);

CREATE INDEX ix_test ON index_test ((
    CASE WHEN flag THEN value1 ELSE value2 END
));
