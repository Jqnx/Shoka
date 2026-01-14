ALTER TABLE "favorite_archives" DROP CONSTRAINT "favorite_archives_archive_id_archive_id_fk";
--> statement-breakpoint
ALTER TABLE "reading_progress" DROP CONSTRAINT "reading_progress_archive_id_archive_id_fk";
--> statement-breakpoint
ALTER TABLE "favorite_archives" ADD CONSTRAINT "favorite_archives_archive_id_archive_id_fk" FOREIGN KEY ("archive_id") REFERENCES "public"."archive"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "reading_progress" ADD CONSTRAINT "reading_progress_archive_id_archive_id_fk" FOREIGN KEY ("archive_id") REFERENCES "public"."archive"("id") ON DELETE cascade ON UPDATE no action;