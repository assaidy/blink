-- name: InsertClient :exec
insert into clients (id, platform, os)
values ($1, $2, $3);

-- name: CheckClientID :one
select exists (select 1 from clients where id = $1 for update);
