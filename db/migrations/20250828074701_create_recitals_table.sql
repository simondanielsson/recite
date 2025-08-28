-- migrate:up
CREATE TABLE recitals (
  id SERIAL PRIMARY KEY,
  url TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  status TEXT NOT NULL,
  path TEXT NOT NULL,
  created_at DATE NOT NULL
);

-- migrate:down
DROP TABLE recitals;
