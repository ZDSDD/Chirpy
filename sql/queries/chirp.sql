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
    CASE
        WHEN $1 = 'asc' THEN created_at
        ELSE NULL
    END ASC,
    CASE
        WHEN $1 = 'desc' THEN created_at
        ELSE NULL
    END DESC;


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

-- name: GetChirpsByUser :many
SELECT chirps.*
FROM chirps
WHERE user_id = $1
ORDER BY
    CASE
        WHEN $2 = 'asc' THEN created_at
        ELSE NULL
    END ASC,
    CASE
        WHEN $2 = 'desc' THEN created_at
        ELSE NULL
    END DESC;