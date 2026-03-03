-- name: InsertChatMessage :exec
insert into chat_messages (id, sender_id, receiver_id, content, sent_at)
values ($1, $2, $3, $4, $5);

-- name: GetChats :many
with last_messages as (
  select
    (case when m.sender_id = @user_id then m.receiver_id else m.sender_id end)::varchar as partner_id,
    max(m.id)::varchar as last_message_id
  from chat_messages m
  where is_deleted = false and (m.sender_id = @user_id or m.receiver_id = @user_id)
  group by partner_id
)
select
  u.id,
  u.name,
  u.username,
  lm.last_message_id
from last_messages lm
join users u on u.id = lm.partner_id
where (@last_message_id_with_last_partner::varchar = '' or lm.last_message_id < @last_message_id_with_last_partner::varchar)
order by lm.last_message_id desc
limit $1;

-- name: CheckChatPartnerID :one
select exists (
  select 1
  from chat_messages
  where is_deleted = false and
        ((sender_id = @user_id and receiver_id = @partner_id) or (sender_id = @partner_id and receiver_id = @user_id))
  for update
);

-- name: GetChatMessages :many
select
  id,
  content,
  sent_at,
  is_read,
  (sender_id = @user_id) as from_me
from chat_messages 
where is_deleted = false and
      ((sender_id = @user_id and receiver_id = @partner_id) or (sender_id = @partner_id and receiver_id = @user_id)) and
      (@last_message_id::varchar = '' or id < @last_message_id::varchar)
order by id desc
limit $1;

-- name: GetChatMessageByID :one
select sender_id, receiver_id
from chat_messages
where is_deleted = false and id = $1
for update;

-- name: MarkMessagesAsRead :many
update chat_messages 
set is_read = true
where receiver_id = @user_id and sender_id = @partner_id and
      (@upto_message_id::varchar = '' or id <= @upto_message_id::varchar) and
      is_read = false
returning id;

-- name: MarkChatMessageAsDeleted :exec
update chat_messages set is_deleted = true where id = $1;

-- name: UpdateChatMessageContent :exec
update chat_messages set content = $1 where id = $2;

-- name: MarkChatAsDeleted :exec
update chat_messages
set is_deleted = true
where ((sender_id = @user_id and receiver_id = @partner_id) or (sender_id = @partner_id and receiver_id = @user_id));

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
      where (case when is_deleted = true then 1 else 0 end) = 1
      limit 1000
    );
    get diagnostics rows_deleted = row_count;
    exit when rows_deleted = 0;
    -- perform pg_sleep(@batch_delay::interval);
  end loop;
end $$;

-- name: GetAllChatPartnerIDs :many
select distinct
  (case when sender_id = @user_id then receiver_id else sender_id end)::varchar as id
from chat_messages
where is_deleted = false and (sender_id = @user_id or receiver_id = @user_id);

-- name: GetUnreadCountsForUser :many
select
  sender_id as partner_id,
  count(*)
from chat_messages
where is_deleted = false and receiver_id = @user_id and is_read = false
group by sender_id;
