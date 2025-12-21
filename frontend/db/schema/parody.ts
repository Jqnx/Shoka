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

export const parody = pgTable(
  "parody",
  {
    id: serial("id").primaryKey(),
    name: text("name").notNull(),
    count: integer("count").notNull(),
  },
  (table) => [uniqueIndex("idx_parody").on(table.name)],
);

export const archiveParody = pgTable(
  "archive_parody",
  {
    archiveId: char("archive_id", { length: 8 }).references(() => archive.id, {
      onDelete: "cascade",
    }),
    parodyId: integer("parody_id").references(() => parody.id, {
      onDelete: "cascade",
    }),
  },
  (table) => [primaryKey({ columns: [table.archiveId, table.parodyId] })],
);
