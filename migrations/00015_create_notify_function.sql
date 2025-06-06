-- +goose Up
-- +goose StatementBegin
create function notify_archives()
returns trigger
as $$
DECLARE
BEGIN
  PERFORM pg_notify('notifyarchives', to_jsonb(NEW)::text);
  RETURN NEW;
END;
$$
language plpgsql
;

CREATE TRIGGER archive_insert
AFTER INSERT ON archives FOR EACH ROW EXECUTE PROCEDURE notify_archives();

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop trigger archive_insert on archives;
drop function notify_archives()
;

-- +goose StatementEnd


