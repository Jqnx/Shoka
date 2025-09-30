-- +goose Up
-- +goose StatementBegin
create or replace function archives_ts_update_trigger_func()
returns trigger
as $$
begin
    perform update_archive_search_vector(new.id);
    return new;
end;
$$
language plpgsql
;

create trigger archives_ts_update_trigger
after insert or update of title, language, category on archives
for each row execute function archives_ts_update_trigger_func();
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop trigger archives_ts_update_trigger on archives;
drop function archives_ts_update_trigger_func()
;
-- +goose StatementEnd


