--------------------------------------------------
-- +goose Up
--------------------------------------------------
create table users (
  id varchar not null,
  name varchar(50) not null,
  username varchar(50) not null unique,
  email varchar(255) not null unique,
  email_is_verified boolean not null default false,
  bio varchar(255) not null,
  joined_at timestamptz not null default now(),

  primary key (id)
);

create table otps (
  id varchar not null,
  user_id varchar not null,
  otp_hash varchar not null,
  channel varchar not null, -- e.g. email, sms, ...etc
  purpose varchar not null, -- e.g. login, email_verify, ...etc
  created_at timestamptz not null default now(),
  expires_at timestamptz not null,

  primary key (id),
  foreign key (user_id) references users (id) on delete cascade
);

create index on otps (expires_at);

create table sessions (
  id varchar not null,
  token varchar not null unique,
  csrf_token varchar not null unique,
  user_id varchar not null,
  platform varchar not null, -- e.g. IPhone 15 Pro, ThinkPad Z13, Firefox
  os varchar not null, -- e.g. Linux, Android
  created_at timestamptz not null default now(),
  expires_at timestamptz not null,
  -- last_active timestamptz not null default now()

  primary key (id),
  foreign key (user_id) references users (id) on delete cascade
);

create index on sessions (expires_at);

create table chat_messages (
  id varchar not null,
  sender_id varchar not null,
  receiver_id varchar not null,
  content varchar not null,
  sent_at timestamptz not null default now(),
  is_read boolean not null default false,
  is_deleted boolean not null default false,

  primary key (id),
  foreign key (sender_id) references users (id) on delete set null,
  foreign key (receiver_id) references users (id) on delete set null
);

create index on chat_messages (sender_id, receiver_id);
create index on chat_messages ((case when is_deleted = true then 1 else 0 end));

--------------------------------------------------
-- +goose Down
--------------------------------------------------
drop table chat_messages;
drop table sessions;
drop table otps;
drop table users;
