-- +goose Up
create table chirps(
    id UUID primary key,
    created_at TIMESTAMP not null,
    updated_at TIMESTAMP not null,
    user_id UUID not null,
    body TEXT not null,
    constraint fk_user foreign key (user_id) references users(id) on delete cascade
);

-- +goose Down
drop table chirps;