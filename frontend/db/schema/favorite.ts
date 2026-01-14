import {
  char,
  pgTable,
  primaryKey,
  timestamp,
  uuid,
} from "drizzle-orm/pg-core";
import { archive } from "./archive";
import { user } from "./better-auth";

export const favoriteArchives = pgTable(
  "favorite_archives",
  {
    archiveId: char("archive_id", { length: 8 }).references(() => archive.id, {
      onDelete: "cascade",
    }),
    userId: uuid("user_id").references(() => user.id),
    favoritedAt: timestamp("favorited_at", { withTimezone: true }).notNull(),
  },
  (table) => [primaryKey({ columns: [table.archiveId, table.userId] })],
);
