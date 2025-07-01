-- +goose Up
-- +goose StatementBegin
create or replace function characters_ts_update_trigger_func()
returns trigger
as
    $$
DECLARE
    r RECORD;
BEGIN
    IF OLD.name IS DISTINCT FROM NEW.name THEN
        FOR r IN SELECT archive_id FROM archives_characters WHERE character_id = NEW.id LOOP
            PERFORM update_archive_search_vector(r.archive_id);
        END LOOP;
    END IF;
    RETURN NEW;
END;
$$
language plpgsql
;

CREATE TRIGGER characters_ts_update_trigger
AFTER UPDATE OF name ON characters
FOR EACH ROW EXECUTE FUNCTION characters_ts_update_trigger_func();
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop trigger characters_ts_update_trigger on characters;
drop function characters_ts_update_trigger_func()
;
-- +goose StatementEnd


