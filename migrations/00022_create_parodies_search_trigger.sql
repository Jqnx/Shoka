-- +goose Up
-- +goose StatementBegin
create or replace function parodies_ts_update_trigger_func()
returns trigger
as $$
DECLARE
    r RECORD;
BEGIN
    IF OLD.parody IS DISTINCT FROM NEW.parody THEN
        FOR r IN SELECT archive_id FROM archives_parodies WHERE parody_id = NEW.id LOOP
            PERFORM update_archive_search_vector(r.archive_id);
        END LOOP;
    END IF;
    RETURN NEW;
END;
$$
language plpgsql
;

CREATE TRIGGER parodies_ts_update_trigger
AFTER UPDATE OF parody ON parodies
FOR EACH ROW EXECUTE FUNCTION parodies_ts_update_trigger_func();
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop trigger parodies_ts_update_trigger on parodies;
drop function parodies_ts_update_trigger_func()
;
-- +goose StatementEnd


