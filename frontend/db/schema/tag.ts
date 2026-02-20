import {
  char,
  integer,
  pgTable,
  primaryKey,
  serial,
  text,
  uniqueIndex,
} from "drizzle-orm/pg-core";
import { archive } from "./archive";

export const tag = pgTable(
  "tag",
  {
    id: serial("id").primaryKey(),
    name: text("name").notNull(),
    description: text("description"),
    count: integer("count").notNull(),
  },
  (table) => [uniqueIndex("idx_tag").on(table.name)],
);

export const archiveTag = pgTable(
  "archive_tag",
  {
    archiveId: char("archive_id", { length: 8 }).references(() => archive.id, {
      onDelete: "cascade",
    }),
    tagId: integer("tag_id").references(() => tag.id, {
      onDelete: "cascade",
    }),
  },
  (table) => [primaryKey({ columns: [table.archiveId, table.tagId] })],
);
