-- name: InsertSession :exec
insert into sessions (id, token, csrf_token, user_id, os, platform, expires_at)
values ($1, $2, $3, $4, $5, $6, $7);

-- name: GetSessionById :one
select * from sessions where id = $1;

-- name: CheckSessionForUser :one
select exists (select 1 from sessions where id = $1 and user_id = $2 for update);

-- name: CheckCsrfTokenForSession :one
select exists (select 1 from sessions where id = @session_id and csrf_token = $1 for update);

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
  id,
  platform, 
  os
from sessions s
where user_id = $1;
