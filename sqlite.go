package gotosql

import (
	"errors"
	"strings"
)

type sqliteGenerator struct{}

func newSQLiteGenerator() SQLGen {
	return &sqliteGenerator{}
}

func (g *sqliteGenerator) GetAutoIncrementKey() string {
	return "AUTOINCREMENT"
}

func (g *sqliteGenerator) GetDefaultValue(sqlType string) (string, error) {
	switch sqlType {
	case "INTEGER":
		return "0", nil
	case "NULL":
		return "", nil
	case "REAL":
		return "0.0", nil
	case "TEXT":
		return "''", nil
	}

	return "", errors.New("unrecognized type " + sqlType)
}

func (g *sqliteGenerator) GenSQLType(goType string) (string, error) {
	switch goType {
	case "[]byte":
		return "BLOB", nil
	case "bool", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "INTEGER", nil
	case "float32", "float64":
		return "REAL", nil
	case "string", "time.Time":
		return "TEXT", nil
	}

	return "", errors.New("unknown type " + goType)
}

func (g *sqliteGenerator) ValidateType(sqlType string) bool {
	switch strings.ToUpper(sqlType) {
	case "BLOB", "INTEGER", "NULL", "REAL", "TEXT":
		return true
	}

	return false
}
