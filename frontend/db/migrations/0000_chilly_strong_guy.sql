CREATE TABLE "archive" (
	"id" char(8) PRIMARY KEY NOT NULL,
	"title" text NOT NULL,
	"summary" text,
	"language" char(2),
	"category" text,
	"page_count" smallint NOT NULL,
	"file_path" text NOT NULL,
	"file_hash" varchar(64) NOT NULL,
	"thumb_path" text,
	"release_date" timestamp,
	"created_at" timestamp NOT NULL,
	"updated_at" timestamp NOT NULL,
	CONSTRAINT "archive_file_path_unique" UNIQUE("file_path"),
	CONSTRAINT "archive_file_hash_unique" UNIQUE("file_hash")
);
--> statement-breakpoint
CREATE TABLE "archive_artist" (
	"archive_id" char(8),
	"artist_id" integer,
	CONSTRAINT "archive_artist_archive_id_artist_id_pk" PRIMARY KEY("archive_id","artist_id")
);
--> statement-breakpoint
CREATE TABLE "artist" (
	"id" serial PRIMARY KEY NOT NULL,
	"name" text NOT NULL,
	"count" integer NOT NULL
);
--> statement-breakpoint
CREATE TABLE "artist_alias" (
	"id" serial PRIMARY KEY NOT NULL,
	"artist_id" integer NOT NULL,
	"alias" text NOT NULL,
	CONSTRAINT "artist_alias_alias_unique" UNIQUE("alias")
);
--> statement-breakpoint
CREATE TABLE "artist_url" (
	"id" serial PRIMARY KEY NOT NULL,
	"artist_id" integer NOT NULL,
	"url" text NOT NULL,
	CONSTRAINT "artist_url_url_unique" UNIQUE("url")
);
--> statement-breakpoint
CREATE TABLE "account" (
	"id" text PRIMARY KEY NOT NULL,
	"account_id" text NOT NULL,
	"provider_id" text NOT NULL,
	"user_id" text NOT NULL,
	"access_token" text,
	"refresh_token" text,
	"id_token" text,
	"access_token_expires_at" timestamp,
	"refresh_token_expires_at" timestamp,
	"scope" text,
	"password" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp NOT NULL
);
--> statement-breakpoint
CREATE TABLE "jwks" (
	"id" text PRIMARY KEY NOT NULL,
	"public_key" text NOT NULL,
	"private_key" text NOT NULL,
	"created_at" timestamp NOT NULL,
	"expires_at" timestamp
);
--> statement-breakpoint
CREATE TABLE "session" (
	"id" text PRIMARY KEY NOT NULL,
	"expires_at" timestamp NOT NULL,
	"token" text NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp NOT NULL,
	"ip_address" text,
	"user_agent" text,
	"user_id" text NOT NULL,
	CONSTRAINT "session_token_unique" UNIQUE("token")
);
--> statement-breakpoint
CREATE TABLE "user" (
	"id" text PRIMARY KEY NOT NULL,
	"name" text NOT NULL,
	"email" text NOT NULL,
	"email_verified" boolean DEFAULT false NOT NULL,
	"image" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	"username" text,
	"display_username" text,
	CONSTRAINT "user_email_unique" UNIQUE("email"),
	CONSTRAINT "user_username_unique" UNIQUE("username")
);
--> statement-breakpoint
CREATE TABLE "verification" (
	"id" text PRIMARY KEY NOT NULL,
	"identifier" text NOT NULL,
	"value" text NOT NULL,
	"expires_at" timestamp NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL
);
--> statement-breakpoint
CREATE TABLE "archive_character" (
	"archive_id" char(8),
	"character_id" integer,
	CONSTRAINT "archive_character_archive_id_character_id_pk" PRIMARY KEY("archive_id","character_id")
);
--> statement-breakpoint
CREATE TABLE "character" (
	"id" serial PRIMARY KEY NOT NULL,
	"name" text NOT NULL,
	"count" integer NOT NULL
);
--> statement-breakpoint
CREATE TABLE "favorite_archives" (
	"archive_id" char(8),
	"user_id" text,
	"favorited_at" timestamp NOT NULL,
	CONSTRAINT "favorite_archives_archive_id_user_id_pk" PRIMARY KEY("archive_id","user_id")
);
--> statement-breakpoint
CREATE TABLE "archive_parody" (
	"archive_id" char(8),
	"parody_id" integer,
	CONSTRAINT "archive_parody_archive_id_parody_id_pk" PRIMARY KEY("archive_id","parody_id")
);
--> statement-breakpoint
CREATE TABLE "parody" (
	"id" serial PRIMARY KEY NOT NULL,
	"name" text NOT NULL,
	"count" integer NOT NULL
);
--> statement-breakpoint
CREATE TABLE "reading_progress" (
	"archive_id" char(8),
	"user_id" text,
	"page" smallint NOT NULL,
	"status" text NOT NULL,
	"last_read" timestamp NOT NULL,
	CONSTRAINT "reading_progress_archive_id_user_id_pk" PRIMARY KEY("archive_id","user_id")
);
--> statement-breakpoint
CREATE TABLE "archive_tag" (
	"archive_id" char(8),
	"tag_id" integer,
	CONSTRAINT "archive_tag_archive_id_tag_id_pk" PRIMARY KEY("archive_id","tag_id")
);
--> statement-breakpoint
CREATE TABLE "tag" (
	"id" serial PRIMARY KEY NOT NULL,
	"name" text NOT NULL,
	"count" integer NOT NULL
);
--> statement-breakpoint
CREATE TABLE "archive_url" (
	"id" serial PRIMARY KEY NOT NULL,
	"archive_id" char(8),
	"url" text NOT NULL
);
--> statement-breakpoint
ALTER TABLE "archive_artist" ADD CONSTRAINT "archive_artist_archive_id_archive_id_fk" FOREIGN KEY ("archive_id") REFERENCES "public"."archive"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "archive_artist" ADD CONSTRAINT "archive_artist_artist_id_artist_id_fk" FOREIGN KEY ("artist_id") REFERENCES "public"."artist"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "artist_alias" ADD CONSTRAINT "artist_alias_artist_id_artist_id_fk" FOREIGN KEY ("artist_id") REFERENCES "public"."artist"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "artist_url" ADD CONSTRAINT "artist_url_artist_id_artist_id_fk" FOREIGN KEY ("artist_id") REFERENCES "public"."artist"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "account" ADD CONSTRAINT "account_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "session" ADD CONSTRAINT "session_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "archive_character" ADD CONSTRAINT "archive_character_archive_id_archive_id_fk" FOREIGN KEY ("archive_id") REFERENCES "public"."archive"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "archive_character" ADD CONSTRAINT "archive_character_character_id_character_id_fk" FOREIGN KEY ("character_id") REFERENCES "public"."character"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "favorite_archives" ADD CONSTRAINT "favorite_archives_archive_id_archive_id_fk" FOREIGN KEY ("archive_id") REFERENCES "public"."archive"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "favorite_archives" ADD CONSTRAINT "favorite_archives_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "archive_parody" ADD CONSTRAINT "archive_parody_archive_id_archive_id_fk" FOREIGN KEY ("archive_id") REFERENCES "public"."archive"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "archive_parody" ADD CONSTRAINT "archive_parody_parody_id_parody_id_fk" FOREIGN KEY ("parody_id") REFERENCES "public"."parody"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "reading_progress" ADD CONSTRAINT "reading_progress_archive_id_archive_id_fk" FOREIGN KEY ("archive_id") REFERENCES "public"."archive"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "reading_progress" ADD CONSTRAINT "reading_progress_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "archive_tag" ADD CONSTRAINT "archive_tag_archive_id_archive_id_fk" FOREIGN KEY ("archive_id") REFERENCES "public"."archive"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "archive_tag" ADD CONSTRAINT "archive_tag_tag_id_tag_id_fk" FOREIGN KEY ("tag_id") REFERENCES "public"."tag"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "archive_url" ADD CONSTRAINT "archive_url_archive_id_archive_id_fk" FOREIGN KEY ("archive_id") REFERENCES "public"."archive"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
CREATE INDEX "idx_title" ON "archive" USING btree ("title");--> statement-breakpoint
CREATE UNIQUE INDEX "idx_file_hash" ON "archive" USING btree ("file_hash");--> statement-breakpoint
CREATE UNIQUE INDEX "idx_thumb_path" ON "archive" USING btree ("thumb_path");--> statement-breakpoint
CREATE UNIQUE INDEX "idx_artist_name" ON "artist" USING btree ("name");--> statement-breakpoint
CREATE INDEX "account_userId_idx" ON "account" USING btree ("user_id");--> statement-breakpoint
CREATE INDEX "session_userId_idx" ON "session" USING btree ("user_id");--> statement-breakpoint
CREATE INDEX "verification_identifier_idx" ON "verification" USING btree ("identifier");--> statement-breakpoint
CREATE UNIQUE INDEX "idx_character" ON "character" USING btree ("name");--> statement-breakpoint
CREATE UNIQUE INDEX "idx_parody" ON "parody" USING btree ("name");--> statement-breakpoint
CREATE UNIQUE INDEX "idx_tag" ON "tag" USING btree ("name");