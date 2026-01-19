-- name: InsertUser :exec
insert into users (id, name, username, email, bio)
values ($1, $2, $3, $4, $5);

-- name: CheckUserID :one
select exists (select 1 from users where id = $1 for update);

-- name: CheckUsername :one
select exists (select 1 from users where username = $1 for update);

-- name: CheckEmail :one
select exists (select 1 from users where email = $1 for update);

-- name: GetUserByUsername :one
select * from users where username = $1 for update;

-- name: GetUserByEmail :one
select * from users where email = $1 for update;

-- name: MarkEmailAsVerified :exec
update users set email_is_verified = true where id = $1;

-- name: GetUserByID :one
select * from users where id = $1 for update;

-- name: UpdateUser :exec
update users set 
    name = $1,
    username = $2,
    email = $3,
    bio = $4
where id = $5;

-- name: DeleteUser :exec
delete from users where id = $1;

-- name: SearchUsers :many
select *
from users
where (name ilike '%' || @query::varchar || '%' or username ilike '%' || @query::varchar || '%') and 
      (@last_id::varchar = '' or id < @last_id::varchar)
order by id desc
limit $1;
