-- +goose Up
-- +goose StatementBegin
create or replace function archives_ts_update_trigger_func()
returns trigger
as $$
BEGIN
    PERFORM update_archive_search_vector(NEW.id);
    RETURN NEW;
END;
$$
language plpgsql
;

CREATE TRIGGER archives_ts_update_trigger
AFTER INSERT OR UPDATE OF title, language, category ON archives
FOR EACH ROW EXECUTE FUNCTION archives_ts_update_trigger_func();
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop trigger archives_ts_update_trigger on archives;
drop function archives_ts_update_trigger_func()
;
-- +goose StatementEnd


