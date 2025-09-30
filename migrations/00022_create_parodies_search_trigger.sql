-- +goose Up
-- +goose StatementBegin
create or replace function parodies_ts_update_trigger_func()
returns trigger
as $$
declare
    r record;
begin
    if old.name is distinct from new.name then
        for r in select archive_id from archives_parodies where parody_id = new.id loop
            perform update_archive_search_vector(r.archive_id);
        end loop;
    end if;
    return new;
end;
$$
language plpgsql
;

create trigger parodies_ts_update_trigger
after update of name on parodies
for each row execute function parodies_ts_update_trigger_func();
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop trigger parodies_ts_update_trigger on parodies;
drop function parodies_ts_update_trigger_func()
;
-- +goose StatementEnd


