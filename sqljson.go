package sqljson

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// Field of T is a field whose value is stored in JSON format in the database.
type Field[T any] struct {
	data T
}

func As[T any](data T) *Field[T] {
	return &Field[T]{data: data}
}

func (pc Field[T]) Get() T {
	return pc.data
}

func (pc *Field[T]) Scan(val any) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &pc.data)
		return nil
	case string:
		json.Unmarshal([]byte(v), &pc.data)
		return nil
	case nil:
		return nil
	default:
		return errors.Errorf("unsupported unmarshal type in Field.Scan: %T", v)
	}
}

const jsonNull = "null"

var (
	_ sql.Scanner   = (*Field[any])(nil)
	_ driver.Valuer = (*Field[any])(nil)
)

func (pc *Field[T]) Value() (driver.Value, error) {
	buf, err := json.Marshal(pc.data)
	if err != nil {
		return nil, errors.Wrap(err, "when marshaling json in Field")
	}
	// pgx breaks when trying to put []byte into jsonb, go figure?
	ret := string(buf)
	if ret == jsonNull {
		return nil, nil
	}
	return ret, nil
}
