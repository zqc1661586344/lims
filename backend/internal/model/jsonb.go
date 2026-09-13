package model

import (
	"bytes"
	"database/sql/driver"
	"fmt"
)

type JSONB []byte

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
	case string:
		*j = JSONB(v)
	default:
		return fmt.Errorf("JSONB.Scan: unsupported type %T", value)
	}
	return nil
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("JSONB.UnmarshalJSON: receiver is nil")
	}
	if bytes.Equal(data, []byte("null")) {
		*j = nil
		return nil
	}
	*j = append((*j)[0:0], data...)
	return nil
}

func (j JSONB) IsEmpty() bool {
	return len(j) == 0 || bytes.Equal(j, []byte("null"))
}

func (j JSONB) String() string {
	return string(j)
}
