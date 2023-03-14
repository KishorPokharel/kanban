create extension if not exists "citext";
create type categorytype as enum ('TODO', 'IN PROGRESS', 'TESTING', 'DONE');

create table users (
    id bigserial primary key,
    username text not null,
    email citext not null unique,
    password bytea not null,
    created_at timestamp(0) with time zone not null default now()
);

create table if not exists tokens (
    hash bytea primary key,
    user_id bigint not null references users(id),
    expiry timestamp(0) with time zone not null
);

create table tasks (
    id bigserial primary key,
    user_id bigint not null references users(id),
    content text not null,
    category categorytype not null default 'TODO',
    created_at timestamp(0) with time zone not null default now()
);

create table taskorder (
    user_id bigint not null references users(id),
    category categorytype not null,
    value bigint[]
);
