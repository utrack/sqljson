package sqljson

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

type jsonField[T any] struct {
	Data T
}

func As[T any](data T) *jsonField[T] {
	return &jsonField[T]{Data: data}
}

func (pc *jsonField[T]) Scan(val any) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, pc.Data)
		return nil
	case string:
		json.Unmarshal([]byte(v), pc.Data)
		return nil
	case nil:
		return nil
	default:
		return errors.Errorf("unsupported unmarshal type in jsonField.Scan: %T", v)
	}
}

const jsonNull = "null"

var (
	_ sql.Scanner   = (*jsonField[any])(nil)
	_ driver.Valuer = (*jsonField[any])(nil)
)

func (pc *jsonField[T]) Value() (driver.Value, error) {
	buf, err := json.Marshal(pc.Data)
	if err != nil {
		return nil, errors.Wrap(err, "when marshaling json in jsonField")
	}
	// pgx breaks when trying to put []byte into jsonb, go figure?
	ret := string(buf)
	if ret == jsonNull {
		return nil, nil
	}
	return ret, nil
}
