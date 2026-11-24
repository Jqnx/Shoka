-- +goose Up
-- A blank language means "unset", but the language facet only excluded NULL,
-- so any stored '' surfaced as a blank option that collided with the editors'
-- own "None" entry. Normalize existing rows; GetAllLanguage also excludes
-- blanks now, so a value written by an older client stays hidden.
UPDATE archive SET language = NULL WHERE trim(language) = '';
UPDATE audio SET language = NULL WHERE trim(language) = '';

-- +goose Down
-- Irreversible on purpose: the original value was blank, which means the same
-- unset state as NULL everywhere language is read.
SELECT 1;
