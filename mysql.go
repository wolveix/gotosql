package gotosql

import (
	"errors"
	"strings"
)

type mysqlGenerator struct{}

func newMySQLGenerator() SQLGen {
	return &mysqlGenerator{}
}

func (g *mysqlGenerator) GetAutoIncrementKey() string {
	return "AUTO_INCREMENT"
}

func (g *mysqlGenerator) GetDefaultValue(sqlType string) (string, error) {
	sqlType = strings.Split(stripBrackets(sqlType), " ")[0]

	switch sqlType {
	case "BIGINT", "INT", "INTEGER", "TINYINT":
		return "0", nil
	case "DATETIME":
		return "CURRENT_TIMESTAMP", nil
	case "DOUBLE", "FLOAT":
		return "0.0", nil
	case "NULL":
		return "", nil
	case "TEXT", "VARCHAR":
		return "''", nil
	}

	return "", errors.New("unrecognized type " + sqlType)
}

func (g *mysqlGenerator) GenSQLType(goType string) (string, error) {
	switch goType {
	case "int64":
		return "BIGINT", nil
	case "uint64":
		return "BIGINT UNSIGNED", nil
	case "[]byte":
		return "BLOB", nil
	case "time.Time":
		return "DATETIME", nil
	case "float64":
		return "DOUBLE", nil
	case "float32":
		return "FLOAT", nil
	case "int", "int8", "int16", "int32":
		return "INT", nil
	case "uint", "uint8", "uint16", "uint32":
		return "INT UNSIGNED", nil
	case "string":
		return "VARCHAR(255)", nil
	case "bool":
		return "TINYINT(1)", nil
	}

	return "", errors.New("unknown type " + goType)
}

func (g *mysqlGenerator) ValidateType(sqlType string) bool {
	sqlType = stripBrackets(sqlType)

	switch strings.ToUpper(sqlType) {
	case "BIGINT", "BIGINT UNSIGNED", "BLOB", "DATETIME", "DOUBLE", "FLOAT", "INT", "INT UNSIGNED", "INTEGER", "INTEGER UNSIGNED", "NULL", "TEXT", "VARCHAR":
		return true
	}

	return false
}
