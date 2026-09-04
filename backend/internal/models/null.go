package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Null* wrap the standard database/sql Null* types with JSON
// marshaling/unmarshaling that emits the underlying flat value (or `null`)
// instead of Go's default `{"String":"x","Valid":true}` struct shape —
// database/sql's Null* types do not implement json.Marshaler as of Go 1.24
// (verified: json.Marshal(sql.NullString{...}) produces the nested struct
// shape), which would otherwise leak straight into every JSON response that
// returns a model/query-result struct without manually unwrapping each
// nullable field first. Every handler in this codebase that responds with a
// raw struct (rather than a hand-built map) relies on this flat shape.
//
// Field names intentionally mirror the wrapped sql.Null* type exactly
// (String/Int64/Time/Float64 + Valid) so every existing `.String`, `.Time`,
// `.Int64`, `.Float64`, `.Valid` field access across models.go and the
// handlers package keeps compiling unchanged — this is a drop-in swap.
// Scan/Value delegate to the stdlib type, so sqlx `db:"..."` scanning and
// driver value conversion behave identically to sql.NullString etc.

type NullString struct {
	String string
	Valid  bool
}

func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.String)
}

func (n *NullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.String, n.Valid = "", false
		return nil
	}
	if err := json.Unmarshal(data, &n.String); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

func (n *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}
	n.String, n.Valid = s.String, s.Valid
	return nil
}

func (n NullString) Value() (driver.Value, error) {
	return sql.NullString{String: n.String, Valid: n.Valid}.Value()
}

type NullInt64 struct {
	Int64 int64
	Valid bool
}

func (n NullInt64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Int64)
}

func (n *NullInt64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Int64, n.Valid = 0, false
		return nil
	}
	if err := json.Unmarshal(data, &n.Int64); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

func (n *NullInt64) Scan(value interface{}) error {
	var s sql.NullInt64
	if err := s.Scan(value); err != nil {
		return err
	}
	n.Int64, n.Valid = s.Int64, s.Valid
	return nil
}

func (n NullInt64) Value() (driver.Value, error) {
	return sql.NullInt64{Int64: n.Int64, Valid: n.Valid}.Value()
}

type NullFloat64 struct {
	Float64 float64
	Valid   bool
}

func (n NullFloat64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Float64)
}

func (n *NullFloat64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Float64, n.Valid = 0, false
		return nil
	}
	if err := json.Unmarshal(data, &n.Float64); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

func (n *NullFloat64) Scan(value interface{}) error {
	var s sql.NullFloat64
	if err := s.Scan(value); err != nil {
		return err
	}
	n.Float64, n.Valid = s.Float64, s.Valid
	return nil
}

func (n NullFloat64) Value() (driver.Value, error) {
	return sql.NullFloat64{Float64: n.Float64, Valid: n.Valid}.Value()
}

type NullTime struct {
	Time  time.Time
	Valid bool
}

func (n NullTime) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Time)
}

func (n *NullTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Time, n.Valid = time.Time{}, false
		return nil
	}
	if err := json.Unmarshal(data, &n.Time); err != nil {
		return err
	}
	n.Valid = true
	return nil
}

func (n *NullTime) Scan(value interface{}) error {
	var s sql.NullTime
	if err := s.Scan(value); err != nil {
		return err
	}
	n.Time, n.Valid = s.Time, s.Valid
	return nil
}

func (n NullTime) Value() (driver.Value, error) {
	return sql.NullTime{Time: n.Time, Valid: n.Valid}.Value()
}
