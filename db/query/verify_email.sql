-- name: CreateVerifyEmail :one

INSERT INTO verify_emails(
    id,
    username,
    email,
    secret_code
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;
