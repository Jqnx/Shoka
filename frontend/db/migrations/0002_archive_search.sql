-- Custom SQL migration file, put your code below! --
ALTER TABLE "archive" ADD COLUMN "search_vector" tsvector;
-- > statement-breakpoint
create or replace function update_archive_search_vector(p_archive_id char(8))
returns void
as
    $$
DECLARE
    v_archive_title TEXT;
    v_archive_language TEXT;
    v_archive_category TEXT;
    v_tag_string TEXT;
    v_artist_string TEXT;
    v_character_string TEXT;
    v_parody_string TEXT;
BEGIN
    SELECT title, language, category INTO v_archive_title, v_archive_language, v_archive_category
    FROM archive WHERE id = p_archive_id;

    SELECT string_agg(t.name, ' ')
    INTO v_tag_string
    FROM tag t
    JOIN archive_tag arg ON t.id = arg.tag_id
    WHERE arg.archive_id = p_archive_id;

    SELECT string_agg(a.name, ' ')
    INTO v_artist_string
    FROM artist a
    JOIN archive_artist ara ON a.id = ara.artist_id
    WHERE ara.archive_id = p_archive_id;

    SELECT string_agg(c.name, ' ')
    INTO v_character_string
    FROM character c
    JOIN archive_character arc ON c.id = arc.character_id
    WHERE arc.archive_id = p_archive_id;

    SELECT string_agg(p.name, ' ')
    INTO v_parody_string
    FROM parody p
    JOIN archive_parody arp ON p.id = arp.parody_id
    WHERE arp.archive_id = p_archive_id;

    UPDATE archive
    SET search_vector = 
        setweight(to_tsvector('english', COALESCE(v_archive_title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(v_archive_language, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(v_archive_category, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(v_tag_string, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(v_artist_string, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(v_character_string, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(v_parody_string, '')), 'B')
    WHERE id = p_archive_id;
END;
$$
language plpgsql
;
-- > statement-breakpoint
create or replace function archive_ts_update_trigger_func()
returns trigger
as $$
BEGIN
    PERFORM update_archive_search_vector(NEW.id);
    RETURN NEW;
END;
$$
language plpgsql
;
-- > statement-breakpoint
CREATE TRIGGER archive_ts_update_trigger
AFTER INSERT OR UPDATE OF title, language, category ON archive
FOR EACH ROW EXECUTE FUNCTION archive_ts_update_trigger_func();

