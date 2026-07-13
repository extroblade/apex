-- Give every existing user without an avatar a random default (served by the
-- static service at /static/avatars/avatar-N.svg). New users get one at register.

UPDATE users
SET avatar_data_url = CONCAT('/media/avatars/avatar-', FLOOR(1 + RAND() * 8), '.svg')
WHERE avatar_data_url IS NULL OR avatar_data_url = '';
