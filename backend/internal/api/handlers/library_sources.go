package handlers

import (
	"Shoka/internal/database/sqlc"
	"encoding/json"
)

// librarySourceResponse maps a stored per-library source row to its API shape.
// Shared by the read endpoint and the update path so the two can't drift.
func librarySourceResponse(row sqlc.LibrarySource) LibrarySourceResponse {
	resp := LibrarySourceResponse{
		Source:            row.Source,
		Enabled:           row.Enabled != 0,
		MagazineBlocklist: unmarshalStringSlice(row.MagazineBlocklist),
		MiscBlocklist:     unmarshalStringSlice(row.MiscBlocklist),
	}

	if row.Cookies != nil {
		resp.Cookies = *row.Cookies
	}

	if row.ApiKey != nil {
		resp.APIKey = *row.ApiKey
	}

	return resp
}

// applyLibrarySourceUpdate overlays the request's present fields onto params,
// which the caller seeds from the library's existing source row. Pointer
// fields are why a partial PATCH can't wipe a stored value the caller never
// mentioned (see UpdateLibrarySourceRequest).
func applyLibrarySourceUpdate(params *sqlc.UpsertLibrarySourceParams, body UpdateLibrarySourceRequest) {
	if body.Enabled != nil {
		params.Enabled = boolToInt(*body.Enabled)
	}

	if body.Cookies != nil {
		params.Cookies = body.Cookies
	}

	if body.APIKey != nil {
		params.ApiKey = body.APIKey
	}

	if body.MagazineBlocklist != nil {
		magazineJSON, _ := json.Marshal(body.MagazineBlocklist)
		params.MagazineBlocklist = string(magazineJSON)
	}

	if body.MiscBlocklist != nil {
		miscJSON, _ := json.Marshal(body.MiscBlocklist)
		params.MiscBlocklist = string(miscJSON)
	}
}

// orEmptyJSONList returns raw, or the empty JSON list when raw is blank, so a
// source row that predates a blocklist column can't be stored unparsable.
func orEmptyJSONList(raw string) string {
	if raw == "" {
		return "[]"
	}

	return raw
}
