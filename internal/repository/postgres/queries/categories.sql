-- name: ListCategories :many
SELECT * FROM categories ORDER BY parent_id NULLS FIRST, name;

-- name: ListParentCategories :many
SELECT * FROM categories WHERE parent_id IS NULL ORDER BY name;

-- name: ListChildCategories :many
SELECT * FROM categories WHERE parent_id = $1 ORDER BY name;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE slug = $1;
