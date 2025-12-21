-- Custom SQL migration file, put your code below! --
create or replace function artist_ts_update_trigger_func()
returns trigger
as $$
DECLARE
    r RECORD;
BEGIN
    IF OLD.name IS DISTINCT FROM NEW.name THEN
        FOR r IN SELECT archive_id FROM archive_artist WHERE artist_id = NEW.id LOOP
            PERFORM update_archive_search_vector(r.archive_id);
        END LOOP;
    END IF;
    RETURN NEW;
END;
$$
language plpgsql
;
-- > statement-breakpoint
CREATE TRIGGER artist_ts_update_trigger
AFTER UPDATE OF name ON artist
FOR EACH ROW EXECUTE FUNCTION artist_ts_update_trigger_func();
-- > statement-breakpoint
create or replace function archive_artist_ts_update_trigger_func()
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
-- > statement-breakpoint
CREATE TRIGGER archive_artist_ts_update_trigger
AFTER INSERT OR DELETE ON archive_artist
FOR EACH ROW EXECUTE FUNCTION archive_artist_ts_update_trigger_func();

