-- name: InsertSession :exec
insert into sessions (id, token, csrf_token, user_id, client_id, expires_at)
values ($1, $2, $3, $4, $5, $6);

-- name: GetSessionById :one
select * from sessions where id = $1;

-- name: RemoveSession :exec
delete from sessions where id = $1;

-- name: BatchDeleteExpriredSessions :exec
do $$
declare
  rows_deleted integer;
begin
  loop
  delete from sessions
  where ctid in (
    select ctid
    from sessions
    where expires_at <= now()
    limit 1000
  );
  get diagnostics rows_deleted = row_count;
  exit when rows_deleted = 0;
end loop;
end $$;

-- name: GetActiveSessionsForUser :many
select 
  s.id,
  c.platform, 
  c.os,
  c.app
from sessions s
join clients c on c.id = s.client_id
where user_id = $1;
