import {
  char,
  index,
  pgTable,
  smallint,
  text,
  timestamp,
  uniqueIndex,
  varchar,
} from "drizzle-orm/pg-core";

export const archive = pgTable(
  "archive",
  {
    id: char("id", { length: 8 }).primaryKey(),
    title: text("title").notNull(),
    summary: text("summary"),
    language: char("language", { length: 2 }),
    category: text("category"),
    pageCount: smallint("page_count").notNull(),
    filePath: text("file_path").unique().notNull(),
    fileHash: varchar("file_hash", { length: 64 }).unique().notNull(),
    thumbPath: text("thumb_path"),
    releaseDate: timestamp("release_date"),
    createdAt: timestamp("created_at").notNull(),
    updatedAt: timestamp("updated_at").notNull(),
  },
  (table) => [
    index("idx_title").on(table.title),
    uniqueIndex("idx_file_hash").on(table.fileHash),
    uniqueIndex("idx_thumb_path").on(table.thumbPath),
  ],
);
