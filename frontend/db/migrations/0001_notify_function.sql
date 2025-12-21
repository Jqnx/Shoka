-- Custom SQL migration file, put your code below! ----
create function notify_archives()
returns trigger
as $$
declare
begin
  perform pg_notify('notifyarchives', to_jsonb(new)::text);
  return new;
end;
$$
language plpgsql
;

-- > statement-breakpoint
create trigger archive_insert
after insert on archive for each row execute procedure notify_archives();

