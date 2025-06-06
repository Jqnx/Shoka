-- +goose Up
-- +goose StatementBegin
create or replace function tags_ts_update_trigger_func()
returns trigger
as $$
DECLARE
    r RECORD;
BEGIN
    IF OLD.tag IS DISTINCT FROM NEW.tag THEN
        FOR r IN SELECT archive_id FROM archives_tags WHERE tag_id = NEW.id LOOP
            PERFORM update_archive_search_vector(r.archive_id);
        END LOOP;
    END IF;
    RETURN NEW;
END;
$$
language plpgsql
;

CREATE TRIGGER tags_ts_update_trigger
AFTER UPDATE OF tag ON tags
FOR EACH ROW EXECUTE FUNCTION tags_ts_update_trigger_func();
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop trigger tags_ts_update_trigger on tags;
drop function tags_ts_update_trigger_func()
;
-- +goose StatementEnd


