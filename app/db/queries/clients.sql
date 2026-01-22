-- name: InsertClient :exec
insert into clients (id, platform, os, app)
values ($1, $2, $3, $4);

-- name: CheckClientID :one
select exists (select 1 from clients where id = $1 for update);

-- name: UpdateClient :exec
update clients set app = $1 where id = $2;
