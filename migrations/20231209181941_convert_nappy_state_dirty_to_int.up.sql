  -- Add the new integer column
ALTER TABLE entries
ADD COLUMN new_nappy_state_dirty INT DEFAULT 0;

-- Set the initial values based on the existing boolean column
UPDATE entries SET new_nappy_state_dirty = 3;
UPDATE entries SET new_nappy_state_dirty = 0 WHERE nappy_state_dirty = false;

-- Drop the old boolean column
ALTER TABLE entries DROP COLUMN nappy_state_dirty;

-- Rename the new int column back to the old ones name
ALTER TABLE entries
RENAME COLUMN new_nappy_state_dirty TO nappy_state_dirty;
