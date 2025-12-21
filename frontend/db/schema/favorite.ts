import {
  char,
  pgTable,
  primaryKey,
  text,
  timestamp,
} from "drizzle-orm/pg-core";
import { archive } from "./archive";
import { user } from "./better-auth";

export const favoriteArchives = pgTable(
  "favorite_archives",
  {
    archiveId: char("archive_id", { length: 8 }).references(() => archive.id),
    userId: text("user_id").references(() => user.id),
    favoritedAt: timestamp("favorited_at").notNull(),
  },
  (table) => [primaryKey({ columns: [table.archiveId, table.userId] })],
);
