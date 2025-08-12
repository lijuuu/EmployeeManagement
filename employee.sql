-- name: CreateEmployee :one
INSERT INTO employees (id, name, position, salary, hired_date, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;

-- name: GetEmployeeByID :one
SELECT id, name, position, salary, hired_date, created_at, updated_at
FROM employees
WHERE id = $1;

-- name: UpdateEmployee :exec
UPDATE employees
SET name = $1, position = $2, salary = $3, hired_date = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $5;

-- name: DeleteEmployee :exec
DELETE FROM employees
WHERE id = $1;

-- name: ListEmployees :many
SELECT id, name, position, salary, hired_date, created_at, updated_at
FROM employees;