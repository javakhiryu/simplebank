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

-- name: UpdateUser :one
UPDATE users
SET 
full_name = coalesce(sqlc.narg('full_name'), full_name),
email = coalesce(sqlc.narg('email'), email),
is_email_verified = coalesce(sqlc.narg('is_email_verified'), is_email_verified)
WHERE username = sqlc.arg('username')
RETURNING *;