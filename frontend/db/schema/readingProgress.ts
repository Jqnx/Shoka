import {
  char,
  pgTable,
  primaryKey,
  smallint,
  text,
  timestamp,
} from "drizzle-orm/pg-core";
import { archive } from "./archive";
import { user } from "./better-auth";

export const readingProgress = pgTable(
  "reading_progress",
  {
    archiveId: char("archive_id", { length: 8 }).references(() => archive.id),
    userId: text("user_id").references(() => user.id),
    page: smallint("page").notNull(),
    status: text("status").notNull(),
    lastRead: timestamp("last_read").notNull(),
  },
  (table) => [primaryKey({ columns: [table.archiveId, table.userId] })],
);
