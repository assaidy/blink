-- name: InsertOtp :exec
insert into otps (id, user_id, otp_hash, channel, purpose, expires_at)
values ($1, $2, $3, $4, $5, $6);

-- name: GetOtpByID :one
select * from otps where id = $1 for update;

-- name: DeleteOtp :exec
delete from otps where id = $1;

-- name: BatchDeleteExpiredOtps :exec
do $$
declare
  rows_deleted integer;
begin
  loop
    delete from otps
    where ctid in (
      select ctid
      from otps
      where expires_at <= now()
      limit 1000
    );
    get diagnostics rows_deleted = row_count;
    exit when rows_deleted = 0;
  end loop;
end $$;
