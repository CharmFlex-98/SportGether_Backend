package tools

import (
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
)

type Cursor struct {
	ID     *int64 `json:"ID"`
	IsNext bool   `json:"IsNext"`
}

type Filter struct {
	PrevCursor string `json:"prevCursor"`
	NextCursor string `json:"nextCursor"`
	PageSize   int    `json:"pageSize"`
}

func (filter Filter) IsCursorEmpty() bool {
	return !filter.HasNextCursor() && !filter.HasPrevCursor()
}

func (filter Filter) HasNextCursor() bool {
	return filter.NextCursor != ""
}

func (filter Filter) HasPrevCursor() bool {
	return filter.PrevCursor != ""
}

func (filter Filter) DecodeCursor() (error, *Cursor) {
	if filter.IsCursorEmpty() {
		return nil, &Cursor{}
	}

	toDecode := ""

	if filter.HasNextCursor() {
		toDecode = filter.NextCursor
	} else {
		toDecode = filter.PrevCursor
	}
	data, err := base32.StdEncoding.DecodeString(toDecode)
	if err != nil {
		return err, nil
	}

	var cursor *Cursor
	err = json.Unmarshal(data, &cursor)
	if err != nil {
		return err, nil
	}

	if cursor.IsNext != filter.HasNextCursor() {
		return errors.New(fmt.Sprintf("Unmatched cursor, IsNext expected to be %v but is %v", filter.HasNextCursor(), cursor.IsNext)), nil
	}

	return nil, cursor
}

func (filter Filter) Validate(validator *RequestValidator) {
	validateCursor(filter, validator)
}

func validateCursor(filter Filter, validator *RequestValidator) {
	validator.Check(
		!(len(filter.PrevCursor) != 0 && len(filter.NextCursor) != 0),
		"cursor",
		"Cannot provide 2 cursor at the same time")
}
