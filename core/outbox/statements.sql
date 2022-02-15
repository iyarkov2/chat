
CREATE TABLE task (
  id uuid primary key,
  version integer,
  created_at timestamp,
  updated_at timestamp,

  type varchar(255),
  exec_counter integer,
  status integer,

  data bytea
);
