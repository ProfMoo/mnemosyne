-- name: GetGreeting :one
SELECT message
FROM
    greetings
WHERE
    id = 1;
