import { char, pgTable, serial, text } from "drizzle-orm/pg-core";
import { archive } from "./archive";

export const archiveUrl = pgTable("archive_url", {
  id: serial("id").primaryKey(),
  archiveId: char("archive_id", { length: 8 }).references(() => archive.id, {
    onDelete: "cascade",
  }),
  url: text("url").notNull(),
});
