package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

func jsonMarshal(v interface{}) (driver.Value, error) {
	b, err := json.Marshal(v)
	return string(b), err
}

func jsonScan(src, dst interface{}) error {
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	case nil:
		return nil
	default:
		return errors.New("jsonb: 不支持的扫描类型")
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, dst)
}

// 确保 driver.Valuer / sql.Scanner 接口在编译期被满足
var (
	_ driver.Valuer = Int2DArray{}
	_ sql.Scanner   = (*Int2DArray)(nil)
	_ driver.Valuer = BinDefList{}
	_ sql.Scanner   = (*BinDefList)(nil)
)
