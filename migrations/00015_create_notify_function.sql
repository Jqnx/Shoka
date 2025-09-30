-- +goose Up
-- +goose StatementBegin
create function notify_archives()
returns trigger
as $$
declare
begin
  perform pg_notify('notifyarchives', to_jsonb(new)::text);
  return new;
end;
$$
language plpgsql
;

create trigger archive_insert
after insert on archives for each row execute procedure notify_archives();

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop trigger archive_insert on archives;
drop function notify_archives()
;

-- +goose StatementEnd


