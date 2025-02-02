-- name: CreateUser :one
INSERT INTO users (
  username,
  hashed_password,
  full_name,
  email
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: UpdateUserHashedPassword :one
UPDATE users 
SET hashed_password = $1, 
password_changed_at = now()
WHERE username = $2
RETURNING *;