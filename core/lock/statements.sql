
CREATE TABLE lock (
    name varchar(255) PRIMARY KEY,
    version integer,
    expires_at timestamp
);

SELECT name, version, expires_at FROM lock LIMIT 1;

INSERT INTO lock(name, version, expires_at) values ('test', 1, '2021-12-30 04:05:06');