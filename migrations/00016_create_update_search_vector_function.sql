-- +goose Up
-- +goose StatementBegin
create or replace function update_archive_search_vector(p_archive_id char(8))
returns void
as
    $$
declare
    v_archive_title text;
    v_archive_language text;
    v_archive_category text;
    v_tags_string text;
    v_artists_string text;
    v_characters_string text;
    v_parodies_string text;
begin
    select title, language, category into v_archive_title, v_archive_language, v_archive_category
    from archives where id = p_archive_id;

    select string_agg(t.name, ' ')
    into v_tags_string
    from tags t
    join archives_tags arg on t.id = arg.tag_id
    where arg.archive_id = p_archive_id;

    select string_agg(a.name, ' ')
    into v_artists_string
    from artists a
    join archives_artists ara on a.id = ara.artist_id
    where ara.archive_id = p_archive_id;

    select string_agg(c.name, ' ')
    into v_characters_string
    from characters c
    join archives_characters arc on c.id = arc.character_id
    where arc.archive_id = p_archive_id;

    select string_agg(p.name, ' ')
    into v_parodies_string
    from parodies p
    join archives_parodies arp on p.id = arp.parody_id
    where arp.archive_id = p_archive_id;

    update archives
    set search_vector = 
        setweight(to_tsvector('english', coalesce(v_archive_title, '')), 'a') ||
        setweight(to_tsvector('english', coalesce(v_archive_language, '')), 'b') ||
        setweight(to_tsvector('english', coalesce(v_archive_category, '')), 'b') ||
        setweight(to_tsvector('english', coalesce(v_tags_string, '')), 'b') ||
        setweight(to_tsvector('english', coalesce(v_artists_string, '')), 'b') ||
        setweight(to_tsvector('english', coalesce(v_characters_string, '')), 'b') ||
        setweight(to_tsvector('english', coalesce(v_parodies_string, '')), 'b')
    where id = p_archive_id;
end;
$$
language plpgsql
;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop function update_archive_search_vector
;
-- +goose StatementEnd


