-- +goose Up
-- +goose StatementBegin
create or replace function archives_parodies_ts_update_trigger_func()
returns trigger
as $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        PERFORM update_archive_search_vector(NEW.archive_id);
    ELSIF (TG_OP = 'DELETE') THEN
        PERFORM update_archive_search_vector(OLD.archive_id);
    END IF;
    RETURN NULL; -- Result is ignored for AFTER triggers
END;
$$
language plpgsql
;

CREATE TRIGGER archives_parodies_ts_update_trigger
AFTER INSERT OR DELETE ON archives_parodies
FOR EACH ROW EXECUTE FUNCTION archives_parodies_ts_update_trigger_func();
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop trigger archives_parodies_ts_update_trigger on archives_parodies;
drop function archives_parodies_ts_update_trigger_func()
;
-- +goose StatementEnd


