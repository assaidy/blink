-- name: InsertChatMessage :exec
insert into chat_messages (id, sender_id, receiver_id, content, sent_at)
values ($1, @sender_id::varchar, @receiver_id::varchar, $2, $3);

-- name: GetChats :many
with last_messages as (
  select
    case when m.sender_id = @user_id::varchar then m.receiver_id else m.sender_id end as partner_id,
    max(m.id)::varchar as last_message_id
  from chat_messages m
  where m.sender_id = @user_id::varchar or m.receiver_id = @user_id::varchar
  group by partner_id
)
select
  u.id,
  u.name,
  u.username,
  lm.last_message_id
from last_messages lm
join users u on u.id = lm.partner_id
where @last_message_id_with_last_partner::varchar = '' or lm.last_message_id < @last_message_id_with_last_partner::varchar
order by lm.last_message_id desc
limit $1;

-- name: CheckChatPartnerID :one
select exists (select 1
               from chat_messages
               where (sender_id = @user_id::varchar and receiver_id = @partner_id::varchar) or
                     (sender_id = @partner_id::varchar and receiver_id = @user_id::varchar)
               for update);

-- name: GetChatMessages :many
select
  id,
  content,
  sent_at,
  is_read,
  (sender_id = @user_id::varchar) as from_me
from chat_messages 
where ((sender_id = @user_id::varchar and receiver_id = @partner_id::varchar) or
       (sender_id = @partner_id::varchar and receiver_id = @user_id::varchar)) and
      (@last_message_id::varchar = '' or id < @last_message_id::varchar)
order by id desc
limit $1;

-- name: MarkMessagesAsRead :many
update chat_messages 
set is_read = true
where (receiver_id = @user_id::varchar and sender_id = @partner_id::varchar) and
      (@upto_message_id::varchar = '' or id <= @upto_message_id::varchar) and
      is_read = false
returning id;

-- name: MarkChatAsDeleted :exec
update chat_messages
set sender_id = null, receiver_id = null
where (sender_id = @user_id::varchar and receiver_id = @partner_id::varchar) or 
      (sender_id = @partner_id::varchar and receiver_id = @user_id::varchar);

-- name: BatchDeleteChatMessages :exec
do $$
declare
  rows_deleted int;
begin
  loop
    delete from chat_messages
    where ctid in (
      select ctid
      from chat_messages
      where (case when sender_id is null or receiver_id is null then 1 else 0 end) = 1
      limit 1000
    );
    get diagnostics rows_deleted = row_count;
    exit when rows_deleted = 0;
    -- perform pg_sleep(@batch_delay::interval);
  end loop;
end $$;

-- name: GetAllChatPartnerIDs :many
select distinct
  (case when sender_id = @user_id::varchar then receiver_id else sender_id end)::varchar as id
from chat_messages
where (sender_id is not null and receiver_id is not null) and
      (sender_id = @user_id::varchar or receiver_id = @user_id::varchar);

-- name: GetUnreadCountsForUser :many
select
  sender_id::varchar as partner_id,
  count(*)
from chat_messages
where receiver_id = @user_id::varchar and is_read = false
group by sender_id;
