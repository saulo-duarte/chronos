package util

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type LocalDateTime struct {
	time.Time
}

const layout = "2006-01-02T15:04:05"

var saoPauloLocation *time.Location

func init() {
	var err error
	saoPauloLocation, err = time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		saoPauloLocation = time.FixedZone("BRT", -3*60*60)
	}
}

func ToTimePtr(ldt *LocalDateTime) *time.Time {
	if ldt == nil {
		return nil
	}
	t := ldt.Time
	return &t
}

func (ldt *LocalDateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	t, err := time.ParseInLocation(layout, s, saoPauloLocation)
	if err != nil {
		return err
	}
	ldt.Time = t
	return nil
}

func (ldt LocalDateTime) MarshalJSON() ([]byte, error) {
	if ldt.IsZero() {
		return []byte(`null`), nil
	}
	return []byte(`"` + ldt.In(saoPauloLocation).Format(layout) + `"`), nil
}

func (ldt LocalDateTime) Equal(other LocalDateTime) bool {
	return ldt.Time.Equal(other.Time)
}

func (ldt LocalDateTime) Value() (driver.Value, error) {
	if ldt.IsZero() {
		return nil, nil
	}
	return ldt.Time, nil
}

func (ldt *LocalDateTime) Scan(value interface{}) error {
	if value == nil {
		ldt.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		ldt.Time = v
		return nil
	case []byte:
		parsed, err := time.ParseInLocation(layout, string(v), saoPauloLocation)
		if err != nil {
			return err
		}
		ldt.Time = parsed
		return nil
	case string:
		parsed, err := time.ParseInLocation(layout, v, saoPauloLocation)
		if err != nil {
			return err
		}
		ldt.Time = parsed
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into LocalDateTime", value)
	}
}
