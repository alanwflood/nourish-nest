CREATE TABLE IF NOT EXISTS entries (
  id VARCHAR PRIMARY KEY,
  nappy_state_wet BOOLEAN,
  nappy_state_dirty BOOLEAN,
  notes TEXT,
  created_at DATETIME,
  updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS feeds (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  entry_id VARCHAR,
  start_time DATETIME,
  end_time DATETIME,
  side VARCHAR,
  created_at DATETIME,
  updated_at DATETIME,
  FOREIGN KEY (entry_id) REFERENCES entries(id)
);
