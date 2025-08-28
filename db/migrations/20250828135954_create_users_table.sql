-- migrate:up
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  created_at DATE NOT NULL
);

-- migrate:down
DROP TABLE users;
