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

export const character = pgTable(
  "character",
  {
    id: serial("id").primaryKey(),
    name: text("name").notNull(),
    count: integer("count").notNull(),
  },
  (table) => [uniqueIndex("idx_character").on(table.name)],
);

export const archiveCharacter = pgTable(
  "archive_character",
  {
    archiveId: char("archive_id", { length: 8 }).references(() => archive.id, {
      onDelete: "cascade",
    }),
    characterId: integer("character_id").references(() => character.id, {
      onDelete: "cascade",
    }),
  },
  (table) => [primaryKey({ columns: [table.archiveId, table.characterId] })],
);
