-- +goose Up
-- +goose StatementBegin
create or replace function update_archive_search_vector(p_archive_id bigint)
returns void
as
    $$
DECLARE
    v_archive_title TEXT;
    v_archive_language TEXT;
    v_archive_category TEXT;
    v_tags_string TEXT;
    v_artists_string TEXT;
    v_characters_string TEXT;
    v_parodies_string TEXT;
BEGIN
    SELECT title, language, category INTO v_archive_title, v_archive_language, v_archive_category
    FROM archives WHERE id = p_archive_id;

    SELECT string_agg(t.tag, ' ')
    INTO v_tags_string
    FROM tags t
    JOIN archives_tags arg ON t.id = arg.tag_id
    WHERE arg.archive_id = p_archive_id;

    SELECT string_agg(a.name, ' ')
    INTO v_artists_string
    FROM artists a
    JOIN archives_artists ara ON a.id = ara.artist_id
    WHERE ara.archive_id = p_archive_id;

    SELECT string_agg(c.character, ' ')
    INTO v_characters_string
    FROM characters c
    JOIN archives_characters arc ON c.id = arc.character_id
    WHERE arc.archive_id = p_archive_id;

    SELECT string_agg(p.parody, ' ')
    INTO v_parodies_string
    FROM parodies p
    JOIN archives_parodies arp ON p.id = arp.parody_id
    WHERE arp.archive_id = p_archive_id;

    UPDATE archives
    SET search_vector = 
        setweight(to_tsvector('english', COALESCE(v_archive_title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(v_archive_language, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(v_archive_category, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(v_tags_string, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(v_artists_string, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(v_characters_string, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(v_parodies_string, '')), 'B')
    WHERE id = p_archive_id;
END;
$$
language plpgsql
;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop function update_archive_search_vector
;
-- +goose StatementEnd


