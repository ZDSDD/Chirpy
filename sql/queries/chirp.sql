-- name: CreateChirp :one
INSERT INTO
    chirps (id, created_at, updated_at, user_id, body)
VALUES
    (gen_random_uuid(), NOW(), NOW(), $1, $2) RETURNING *;

-- name: GetChirps :many
SELECT
    chirps.*
FROM
    chirps
ORDER BY
    created_at ASC;

-- name: GetChirp :one
SELECT
    chirps.*
FROM
    chirps
WHERE
    id = $1;

-- name: DeleteChirp :exec
DELETE FROM
    chirps
WHERE
    id = $1;