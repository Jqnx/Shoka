import { char, integer, pgTable, primaryKey } from "drizzle-orm/pg-core";
import { archive } from "./archive";
import { artist } from "./artist";

export const archiveArtist = pgTable(
  "archive_artist",
  {
    archiveId: char("archive_id", { length: 8 }).references(() => archive.id, {
      onDelete: "cascade",
    }),
    artistId: integer("artist_id").references(() => artist.id, {
      onDelete: "cascade",
    }),
  },
  (table) => [primaryKey({ columns: [table.archiveId, table.artistId] })],
);
