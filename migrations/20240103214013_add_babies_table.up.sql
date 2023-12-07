CREATE TABLE IF NOT EXISTS babies (
  id VARCHAR PRIMARY KEY,
  user_id VARCHAR NOT NULL,
  gender VARCHAR NOT NULL,
  first_name VARCHAR NOT NULL,
  last_name VARCHAR NOT NULL,
  date_of_birth DATETIME,
  created_at DATETIME,
  updated_at DATETIME,
  FOREIGN KEY(user_id) REFERENCES users(id)
);
