-- +goose Up
-- +goose StatementBegin
-- Same trigram tokenizer as archive_fts (see 00028): substring matching,
-- which also works for CJK titles and names. No backfill - the table is new.
CREATE VIRTUAL TABLE audio_fts USING fts5(
    audio_id UNINDEXED,
    title,
    artists,
    tags,
    circles,
    voice_actors,
    tokenize = 'trigram'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audio_fts_audio_ai AFTER INSERT ON audio BEGIN
    INSERT INTO audio_fts (audio_id, title, artists, tags, circles, voice_actors)
    VALUES (NEW.id, NEW.title, NULL, NULL, NULL, NULL);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audio_fts_audio_au AFTER UPDATE OF title ON audio BEGIN
    UPDATE audio_fts SET title = NEW.title WHERE audio_id = NEW.id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audio_fts_audio_ad AFTER DELETE ON audio BEGIN
    DELETE FROM audio_fts WHERE audio_id = OLD.id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audio_fts_artist_ai AFTER INSERT ON audio_artist BEGIN
    UPDATE audio_fts SET artists = (
        SELECT group_concat(artist.name, ' ') FROM audio_artist JOIN artist ON artist.id = audio_artist.artist_id WHERE audio_artist.audio_id = NEW.audio_id
    ) WHERE audio_id = NEW.audio_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audio_fts_artist_ad AFTER DELETE ON audio_artist BEGIN
    UPDATE audio_fts SET artists = (
        SELECT group_concat(artist.name, ' ') FROM audio_artist JOIN artist ON artist.id = audio_artist.artist_id WHERE audio_artist.audio_id = OLD.audio_id
    ) WHERE audio_id = OLD.audio_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audio_fts_tag_ai AFTER INSERT ON audio_tag BEGIN
    UPDATE audio_fts SET tags = (
        SELECT group_concat(tag.name, ' ') FROM audio_tag JOIN tag ON tag.id = audio_tag.tag_id WHERE audio_tag.audio_id = NEW.audio_id
    ) WHERE audio_id = NEW.audio_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audio_fts_tag_ad AFTER DELETE ON audio_tag BEGIN
    UPDATE audio_fts SET tags = (
        SELECT group_concat(tag.name, ' ') FROM audio_tag JOIN tag ON tag.id = audio_tag.tag_id WHERE audio_tag.audio_id = OLD.audio_id
    ) WHERE audio_id = OLD.audio_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audio_fts_circle_ai AFTER INSERT ON audio_circle BEGIN
    UPDATE audio_fts SET circles = (
        SELECT group_concat(circle.name, ' ') FROM audio_circle JOIN circle ON circle.id = audio_circle.circle_id WHERE audio_circle.audio_id = NEW.audio_id
    ) WHERE audio_id = NEW.audio_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audio_fts_circle_ad AFTER DELETE ON audio_circle BEGIN
    UPDATE audio_fts SET circles = (
        SELECT group_concat(circle.name, ' ') FROM audio_circle JOIN circle ON circle.id = audio_circle.circle_id WHERE audio_circle.audio_id = OLD.audio_id
    ) WHERE audio_id = OLD.audio_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audio_fts_voice_actor_ai AFTER INSERT ON audio_voice_actor BEGIN
    UPDATE audio_fts SET voice_actors = (
        SELECT group_concat(voice_actor.name, ' ') FROM audio_voice_actor JOIN voice_actor ON voice_actor.id = audio_voice_actor.voice_actor_id WHERE audio_voice_actor.audio_id = NEW.audio_id
    ) WHERE audio_id = NEW.audio_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audio_fts_voice_actor_ad AFTER DELETE ON audio_voice_actor BEGIN
    UPDATE audio_fts SET voice_actors = (
        SELECT group_concat(voice_actor.name, ' ') FROM audio_voice_actor JOIN voice_actor ON voice_actor.id = audio_voice_actor.voice_actor_id WHERE audio_voice_actor.audio_id = OLD.audio_id
    ) WHERE audio_id = OLD.audio_id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
-- artist/circle are the only two linked entities that support renaming; tag
-- and voice-actor names are immutable once created.
CREATE TRIGGER audio_fts_artist_rename AFTER UPDATE OF name ON artist BEGIN
    UPDATE audio_fts SET artists = (
        SELECT group_concat(artist.name, ' ') FROM audio_artist JOIN artist ON artist.id = audio_artist.artist_id WHERE audio_artist.audio_id = audio_fts.audio_id
    )
    WHERE audio_id IN (SELECT audio_id FROM audio_artist WHERE artist_id = NEW.id);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER audio_fts_circle_rename AFTER UPDATE OF name ON circle BEGIN
    UPDATE audio_fts SET circles = (
        SELECT group_concat(circle.name, ' ') FROM audio_circle JOIN circle ON circle.id = audio_circle.circle_id WHERE audio_circle.audio_id = audio_fts.audio_id
    )
    WHERE audio_id IN (SELECT audio_id FROM audio_circle WHERE circle_id = NEW.id);
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_circle_rename;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_artist_rename;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_voice_actor_ad;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_voice_actor_ai;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_circle_ad;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_circle_ai;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_tag_ad;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_tag_ai;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_artist_ad;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_artist_ai;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_audio_ad;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_audio_au;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TRIGGER IF EXISTS audio_fts_audio_ai;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS audio_fts;
-- +goose StatementEnd
