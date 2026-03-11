-- DROP DATABASE IF EXISTS apidb;

--  CREATE DATABASE apidb
--      WITH
--      OWNER = postgres
--      ENCODING = 'UTF8'
--      LC_COLLATE = 'en_US.utf8'
--      LC_CTYPE = 'en_US.utf8'
--      LOCALE_PROVIDER = 'libc'
--      TABLESPACE = pg_default
--      CONNECTION LIMIT = -1
--      IS_TEMPLATE = False;

CREATE TABLE statuses (
    id SERIAL PRIMARY KEY,
    encoded_status TEXT NOT NULL
);
