psql -h 127.0.0.1 -U postgres -p 5432 experiment

pg_recvlogical -h 127.0.0.1 -U postgres -p 5432 -d experiment -S slot1 --start
create database experiment;


CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE outbox (
  id uuid primary key,
  version integer,
  created_at timestamp,
  updated_at timestamp,
  message bytea

);

select * from pg_replication_slots;
select * from pg_stat_replication;


select * from outbox;

-- Create replication slot
SELECT * FROM pg_create_logical_replication_slot('slot1', 'test_decoding');
SELECT * FROM pg_create_logical_replication_slot('slot2', 'wal2json');
-- Check replication slot
SELECT slot_name, plugin, slot_type, database, active, restart_lsn, confirmed_flush_lsn FROM pg_replication_slots;
-- Query data from the replication slot
SELECT * FROM pg_logical_slot_get_changes('slot1', NULL, NULL);

-- Drop Replication Slot
SELECT pg_drop_replication_slot('slot1');

insert into outbox (id, created_at, updated_at, message) values (uuid_generate_v4(), now(), now(), '\x00FF00');


-- Two databases experiment
create database experiment2;
CREATE TABLE t2 (
    id uuid primary key,
    message varchar(255)
);

insert into experiment.t1(id, message) values (uuid_generate_v4(), "M1");
insert into experiment2.t2(id, message) values (uuid_generate_v4(), "M2");

insert into experiment2.t2(id, message) values (uuid_generate_v4(), '');