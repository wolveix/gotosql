package gotosql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

const (
	SqlDialectMySql  = SqlDialect("MYSQL")
	SqlDialectSqlite = SqlDialect("SQLITE")
)

type (
	SqlDialect   string
	SqlGenerator struct {
		customTypes   map[string]string
		dialect       SqlDialect
		gen           SqlGen
		nullDefault   bool
		overrideTypes map[string]string
	}
	SqlGen interface {
		GetDefaultValue(sqlType string) (string, error)
		GetSqlType(goType string) (string, error)
		ValidateType(sqlType string) bool
	}
)

func (d *SqlDialect) String() string {
	return string(*d)
}

func NewSqlGenerator(sqlDialect SqlDialect, nullDefault bool, customTypes map[string]string) (*SqlGenerator, error) {
	generator := SqlGenerator{
		customTypes:   customTypes,
		dialect:       sqlDialect,
		nullDefault:   nullDefault,
		overrideTypes: make(map[string]string),
	}

	switch sqlDialect {
	case SqlDialectMySql:
		generator.gen = newMysqlGenerator()
		break
	case SqlDialectSqlite:
		generator.gen = newSqliteGenerator()
		break
	default:
		return nil, errors.New("unknown sql dialect: " + sqlDialect.String())
	}

	// check that the type is valid against the given dialect
	if len(customTypes) > 0 {
		for k, v := range customTypes {
			customTypes[k] = strings.ToUpper(v)
			if !generator.gen.ValidateType(customTypes[k]) {
				return nil, errors.New("unrecognized sql type " + customTypes[k])
			}
		}
	}

	return &generator, nil
}

func (g *SqlGenerator) Generate(object any, history bool, rawTableName ...string) (string, error) {
	obj := reflect.TypeOf(object)

	var err error
	var sqlStatement, sqlStmt []string
	var tableName string

	if len(rawTableName) > 0 {
		tableName = rawTableName[0]
	} else {
		tableName = strings.ToLower(obj.Name())
		if !strings.HasSuffix(tableName, "s") {
			tableName += "s"
		}
	}

	fields, types := g.getFields(obj)

	for i, field := range fields {
		var sqlDefault, sqlType string

		if customType, ok := g.overrideTypes[field]; ok {
			sqlType = customType
		} else if customType, ok = g.customTypes[types[i]]; ok {
			sqlType = customType
		} else {
			sqlType, err = g.gen.GetSqlType(types[i])
			if err != nil {
				return "", fmt.Errorf("failed to find corresponding SQL type for field:%v; type:%v", field, types[i])
			}
		}

		if !g.nullDefault {
			sqlDefault, err = g.gen.GetDefaultValue(sqlType)
			if err != nil {
				return "", fmt.Errorf("failed to find corresponding SQL default for field:%v; type:%v", field, types[i])
			}

			if sqlDefault != "" {
				sqlDefault = " NOT NULL DEFAULT " + sqlDefault
			}
		}

		sqlStmt = append(sqlStmt, "    "+field+" "+sqlType+sqlDefault)
	}

	// create the table
	sqlStatement = append(sqlStatement, "CREATE TABLE IF NOT EXISTS "+tableName+" (\n"+strings.Join(sqlStmt, ",\n")+"\n);\n")

	if history {
		// create the table history
		sqlStatement = append(sqlStatement, "CREATE TABLE IF NOT EXISTS "+tableName+"_history (\n"+strings.Join(sqlStmt, ",\n")+"\n);\n")

		keys := strings.Join(fields, ", ")
		values := strings.Join(fields, ", new.")

		if g.dialect == SqlDialectSqlite {
			sqlStatement = append(sqlStatement, "CREATE TRIGGER IF NOT EXISTS "+tableName+"_audit BEFORE UPDATE ON "+tableName+" BEGIN INSERT INTO "+tableName+"_history (\n    "+
				keys+"\n) VALUES (\n    new."+values+"\n);\nEND;\n")

			sqlStatement = append(sqlStatement, "CREATE TRIGGER IF NOT EXISTS "+tableName+"_audit_first AFTER INSERT ON "+tableName+" BEGIN INSERT INTO "+tableName+"_history (\n    "+
				keys+"\n) VALUES (\n    new."+values+"\n);\nEND;\n")
		} else {
			sqlStatement = append(sqlStatement, "DROP TRIGGER IF EXISTS "+tableName+"_audit;\n")
			sqlStatement = append(sqlStatement, "DROP TRIGGER IF EXISTS "+tableName+"_audit_first;\n")

			sqlStatement = append(sqlStatement, "CREATE TRIGGER "+tableName+"_audit BEFORE UPDATE ON "+tableName+
				" FOR EACH ROW BEGIN\n    INSERT INTO "+tableName+"_history (\n    "+keys+"\n) VALUES (\n    new."+values+"\n);\nEND;\n")

			sqlStatement = append(sqlStatement, "CREATE TRIGGER "+tableName+"_audit_first AFTER INSERT ON "+tableName+
				" FOR EACH ROW BEGIN\n    INSERT INTO "+tableName+"_history (\n    "+keys+"\n) VALUES (\n    new."+values+"\n);\nEND;\n")
		}
	}

	return strings.Join(sqlStatement, "\n"), nil
}

func camelCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	var b strings.Builder
	var nextUpper bool

	b.Grow(len(s))

	for _, r := range s {
		switch {
		case nextUpper:
			b.WriteRune(unicode.ToUpper(r))
			nextUpper = false
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if b.Len() == 0 {
				b.WriteRune(unicode.ToLower(r))
			} else {
				b.WriteRune(r)
			}
		default:
			nextUpper = true
		}
	}

	return b.String()
}

func (g *SqlGenerator) getFields(t reflect.Type) ([]string, []string) {
	var names, types []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		override := field.Tag.Get("gotosql")
		if override != "" {
			g.overrideTypes[camelCase(field.Name)] = override
		}

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			embeddedNames, embeddedTypes := g.getFields(field.Type)
			names, types = append(names, embeddedNames...), append(types, embeddedTypes...)
		} else {
			fieldType := field.Type.String()
			if field.Type.Kind() != reflect.Struct {
				fieldType = field.Type.Kind().String()
			}

			names, types = append(names, camelCase(field.Name)), append(types, fieldType)
		}
	}

	return names, types
}

func stripBrackets(s string) string {
	var result string
	var inBracket bool
	for _, r := range s {
		if r == '(' {
			inBracket = true
		} else if r == ')' {
			inBracket = false
		} else if !inBracket {
			result += string(r)
		}
	}
	return result
}
