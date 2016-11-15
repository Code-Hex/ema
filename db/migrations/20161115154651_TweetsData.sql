
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
create table users (
    id serial primary key,
    user_id varchar(255) not null,
    category varchar(255) not null,
    inserted_at timestamp not null default current_timestamp
);

create table tweets (
	id serial primary key,
    user_id int references users(id),
    tweet_id bigserial not null,
    tweet text not null,
    image_url varchar(255),
    inserted_at timestamp not null default current_timestamp
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
drop table users cascade;
drop table tweets cascade;
