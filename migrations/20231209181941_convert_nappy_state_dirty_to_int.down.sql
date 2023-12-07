-- Rename the new int column back to the old one's name
ALTER TABLE entries
RENAME COLUMN nappy_state_dirty TO new_nappy_state_dirty;

-- Add back the old boolean column
ALTER TABLE entries
ADD COLUMN nappy_state_dirty BOOLEAN;

-- Set the initial values based on the new integer column
UPDATE entries SET nappy_state_dirty = true WHERE new_nappy_state_dirty > 0;
UPDATE entries SET nappy_state_dirty = false WHERE new_nappy_state_dirty = 0;

-- Drop the new integer column
ALTER TABLE entries DROP COLUMN new_nappy_state_dirty;
