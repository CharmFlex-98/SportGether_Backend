package tools

import (
	"errors"
	"fmt"
)

type Cursor struct {
	LastDistance *float64 `json:"lastDistance"`
	IsNext       bool     `json:"IsNext"`
}

type UserFromLocationFilter struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type Filter struct {
	PrevCursor   string                  `json:"prevCursor"`
	NextCursor   string                  `json:"nextCursor"`
	PageSize     int                     `json:"pageSize"`
	FromLocation *UserFromLocationFilter `json:"fromLocation"`
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
	cursor := Cursor{}
	err := DecodeToBase32(&cursor, toDecode)
	if err != nil {
		return err, nil
	}

	if cursor.IsNext != filter.HasNextCursor() {
		return errors.New(fmt.Sprintf("Unmatched cursor, IsNext expected to be %v but is %v", filter.HasNextCursor(), cursor.IsNext)), nil
	}

	return nil, &cursor
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
