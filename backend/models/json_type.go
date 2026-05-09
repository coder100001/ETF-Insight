package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JSONMap map[string]any

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

func (j *JSONMap) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		*j = nil
		return nil
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*j = m
	return nil
}

func (j JSONMap) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]any(j))
}
