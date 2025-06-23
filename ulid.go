package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/oklog/ulid/v2"
)

var ErrEmptyULID = fmt.Errorf("ulid is empty")

// NullULID is a custom type that implements the sql.Scanner and driver.Valuer interfaces.
// this allows us to use the ULID type in our database calls when dealing with nullable values.
type NullULID struct {
	ULID  ulid.ULID
	Valid bool
}

// Scan implements the sql.Scanner interface.
func (n *NullULID) Scan(value any) error {
	if value == nil {
		n.ULID, n.Valid = ulid.ULID{}, false
		return nil
	}
	if err := n.ULID.Scan(value); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullULID) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.ULID.Value()
}

// MarshalJSON implements the json.Marshaler interface.
func (n NullULID) MarshalJSON() ([]byte, error) {
	if !n.Valid || n.ULID == (ulid.ULID{}) {
		return []byte("null"), nil
	}
	return json.Marshal(n.ULID.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (n *NullULID) UnmarshalJSON(data []byte) error {
	var id string
	if err := json.Unmarshal(data, &id); err != nil {
		return err
	}
	if id == "" {
		n.Valid = false
		return nil
	}
	uu, err := ulid.Parse(id)
	if err != nil {
		return err
	}
	if uu == (ulid.ULID{}) {
		return ErrEmptyULID
	}
	n.Valid = true
	n.ULID = uu
	return nil
}

// String implements the fmt.Stringer interface.
func (n NullULID) String() string {
	if !n.Valid {
		return ""
	}
	return n.ULID.String()
}
