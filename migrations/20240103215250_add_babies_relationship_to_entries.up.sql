ALTER TABLE entries RENAME TO entries_old;

CREATE TABLE entries (
  id VARCHAR PRIMARY KEY,
  nappy_state_wet BOOLEAN,
  nappy_state_dirty INT DEFAULT 0,
  notes TEXT,
  created_at DATETIME,
  updated_at DATETIME,
  baby_id VARCHAR,
  FOREIGN KEY (baby_id) REFERENCES babies(id)
);

INSERT INTO entries
SELECT 
  id,
  nappy_state_wet,
  nappy_state_dirty,
  notes,
  created_at,
  updated_at,
  null
FROM entries_old;

DROP TABLE entries_old;
