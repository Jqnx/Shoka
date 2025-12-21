import {
  integer,
  pgTable,
  serial,
  text,
  uniqueIndex,
} from "drizzle-orm/pg-core";

export const artist = pgTable(
  "artist",
  {
    id: serial("id").primaryKey(),
    name: text("name").notNull(),
    count: integer("count").notNull(),
  },
  (table) => [uniqueIndex("idx_artist_name").on(table.name)],
);

export const artistAlias = pgTable("artist_alias", {
  id: serial("id").primaryKey(),
  artistId: integer("artist_id")
    .references(() => artist.id)
    .notNull(),
  alias: text("alias").unique().notNull(),
});

export const artistUrl = pgTable("artist_url", {
  id: serial("id").primaryKey(),
  artistId: integer("artist_id")
    .references(() => artist.id)
    .notNull(),
  url: text("url").unique().notNull(),
});
