-- +goose Up
-- +goose StatementBegin
create or replace function archives_tags_ts_update_trigger_func()
returns trigger
as $$
begin
    if (tg_op = 'insert') then
        perform update_archive_search_vector(new.archive_id);
    elsif (tg_op = 'delete') then
        perform update_archive_search_vector(old.archive_id);
    end if;
    return null; -- result is ignored for after triggers
end;
$$
language plpgsql
;

create trigger archives_tags_ts_update_trigger
after insert or delete on archives_tags
for each row execute function archives_tags_ts_update_trigger_func();
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop trigger archives_tags_ts_update_trigger on archives_tags;
drop function archives_tags_ts_update_trigger_func()
;
-- +goose StatementEnd


