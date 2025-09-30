-- +goose Up
-- +goose StatementBegin
create or replace function tags_ts_update_trigger_func()
returns trigger
as $$
declare
    r record;
begin
    if old.name is distinct from new.name then
        for r in select archive_id from archives_tags where tag_id = new.id loop
            perform update_archive_search_vector(r.archive_id);
        end loop;
    end if;
    return new;
end;
$$
language plpgsql
;

create trigger tags_ts_update_trigger
after update of name on tags
for each row execute function tags_ts_update_trigger_func();
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop trigger tags_ts_update_trigger on tags;
drop function tags_ts_update_trigger_func()
;
-- +goose StatementEnd


