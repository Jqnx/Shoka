-- +goose Up
-- +goose StatementBegin
-- Trigram tokenizer (not unicode61) so search works as substring matching
-- and handles CJK titles/tags correctly - unicode61 tokenizes on
-- whitespace/punctuation, which is useless for Japanese text that has
-- neither between words.
CREATE VIRTUAL TABLE archive_fts USING fts5(
    archive_id UNINDEXED,
    title,
    summary,
    artists,
    tags,
    parodies,
    circles,
    characters,
    category,
    tokenize = 'trigram'
);
-- +goose StatementEnd

-- +goose StatementBegin
-- One-time backfill for archives that existed before this table/its
-- triggers did. Every write from here on is kept in sync by the triggers
-- below instead of any application-level reindex step.
INSERT INTO archive_fts (archive_id, title, summary, artists, tags, parodies, circles, characters, category)
SELECT
    archive.id,
    archive.title,
    archive.summary,
    (SELECT group_concat(artist.name, ' ') FROM archive_artist JOIN artist ON artist.id = archive_artist.artist_id WHERE archive_artist.archive_id = archive.id),
    (SELECT group_concat(tag.name, ' ') FROM archive_tag JOIN tag ON tag.id = archive_tag.tag_id WHERE archive_tag.archive_id = archive.id),
    (SELECT group_concat(parody.name, ' ') FROM archive_parody JOIN parody ON parody.id = archive_parody.parody_id WHERE archive_parody.archive_id = archive.id),
    (SELECT group_concat(circle.name, ' ') FROM archive_circle JOIN circle ON circle.id = archive_circle.circle_id WHERE archive_circle.archive_id = archive.id),
    (SELECT group_concat(character.name, ' ') FROM archive_character JOIN character ON character.id = archive_character.character_id WHERE archive_character.archive_id = archive.id),
    archive.category
FROM archive;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_archive_ai AFTER INSERT ON archive BEGIN
    INSERT INTO archive_fts (archive_id, title, summary, artists, tags, parodies, circles, characters, category)
    VALUES (NEW.id, NEW.title, NEW.summary, NULL, NULL, NULL, NULL, NULL, NEW.category);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_archive_au AFTER UPDATE OF title, summary, category ON archive BEGIN
    UPDATE archive_fts SET title = NEW.title, summary = NEW.summary, category = NEW.category
    WHERE archive_id = NEW.id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_archive_ad AFTER DELETE ON archive BEGIN
    DELETE FROM archive_fts WHERE archive_id = OLD.id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_tag_ai AFTER INSERT ON archive_tag BEGIN
    UPDATE archive_fts SET tags = (
        SELECT group_concat(tag.name, ' ') FROM archive_tag JOIN tag ON tag.id = archive_tag.tag_id WHERE archive_tag.archive_id = NEW.archive_id
    ) WHERE archive_id = NEW.archive_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_tag_ad AFTER DELETE ON archive_tag BEGIN
    UPDATE archive_fts SET tags = (
        SELECT group_concat(tag.name, ' ') FROM archive_tag JOIN tag ON tag.id = archive_tag.tag_id WHERE archive_tag.archive_id = OLD.archive_id
    ) WHERE archive_id = OLD.archive_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_artist_ai AFTER INSERT ON archive_artist BEGIN
    UPDATE archive_fts SET artists = (
        SELECT group_concat(artist.name, ' ') FROM archive_artist JOIN artist ON artist.id = archive_artist.artist_id WHERE archive_artist.archive_id = NEW.archive_id
    ) WHERE archive_id = NEW.archive_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_artist_ad AFTER DELETE ON archive_artist BEGIN
    UPDATE archive_fts SET artists = (
        SELECT group_concat(artist.name, ' ') FROM archive_artist JOIN artist ON artist.id = archive_artist.artist_id WHERE archive_artist.archive_id = OLD.archive_id
    ) WHERE archive_id = OLD.archive_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_parody_ai AFTER INSERT ON archive_parody BEGIN
    UPDATE archive_fts SET parodies = (
        SELECT group_concat(parody.name, ' ') FROM archive_parody JOIN parody ON parody.id = archive_parody.parody_id WHERE archive_parody.archive_id = NEW.archive_id
    ) WHERE archive_id = NEW.archive_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_parody_ad AFTER DELETE ON archive_parody BEGIN
    UPDATE archive_fts SET parodies = (
        SELECT group_concat(parody.name, ' ') FROM archive_parody JOIN parody ON parody.id = archive_parody.parody_id WHERE archive_parody.archive_id = OLD.archive_id
    ) WHERE archive_id = OLD.archive_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_circle_ai AFTER INSERT ON archive_circle BEGIN
    UPDATE archive_fts SET circles = (
        SELECT group_concat(circle.name, ' ') FROM archive_circle JOIN circle ON circle.id = archive_circle.circle_id WHERE archive_circle.archive_id = NEW.archive_id
    ) WHERE archive_id = NEW.archive_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_circle_ad AFTER DELETE ON archive_circle BEGIN
    UPDATE archive_fts SET circles = (
        SELECT group_concat(circle.name, ' ') FROM archive_circle JOIN circle ON circle.id = archive_circle.circle_id WHERE archive_circle.archive_id = OLD.archive_id
    ) WHERE archive_id = OLD.archive_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_character_ai AFTER INSERT ON archive_character BEGIN
    UPDATE archive_fts SET characters = (
        SELECT group_concat(character.name, ' ') FROM archive_character JOIN character ON character.id = archive_character.character_id WHERE archive_character.archive_id = NEW.archive_id
    ) WHERE archive_id = NEW.archive_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_character_ad AFTER DELETE ON archive_character BEGIN
    UPDATE archive_fts SET characters = (
        SELECT group_concat(character.name, ' ') FROM archive_character JOIN character ON character.id = archive_character.character_id WHERE archive_character.archive_id = OLD.archive_id
    ) WHERE archive_id = OLD.archive_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
-- artist/circle are the only two entities that support renaming (see
-- UpdateArtist/UpdateCircle) - tag/parody/character names are immutable
-- once created, so they need no equivalent trigger.
CREATE TRIGGER archive_fts_artist_rename AFTER UPDATE OF name ON artist BEGIN
    UPDATE archive_fts SET artists = (
        SELECT group_concat(artist.name, ' ') FROM archive_artist JOIN artist ON artist.id = archive_artist.artist_id WHERE archive_artist.archive_id = archive_fts.archive_id
    )
    WHERE archive_id IN (SELECT archive_id FROM archive_artist WHERE artist_id = NEW.id);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER archive_fts_circle_rename AFTER UPDATE OF name ON circle BEGIN
    UPDATE archive_fts SET circles = (
        SELECT group_concat(circle.name, ' ') FROM archive_circle JOIN circle ON circle.id = archive_circle.circle_id WHERE archive_circle.archive_id = archive_fts.archive_id
    )
    WHERE archive_id IN (SELECT archive_id FROM archive_circle WHERE circle_id = NEW.id);
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_circle_rename;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_artist_rename;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_character_ad;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_character_ai;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_circle_ad;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_circle_ai;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_parody_ad;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_parody_ai;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_artist_ad;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_artist_ai;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_tag_ad;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_tag_ai;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_archive_ad;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_archive_au;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS archive_fts_archive_ai;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS archive_fts;
-- +goose StatementEnd
