# Go to SQL

Generate SQL schemas from native Go objects, utilizing appropriate SQL data types

### Supported SQL Engines

- MySQL
- SQLite

### Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/wolveix/gotosql"
)

func main() {
	sqlGenerator, err := gotosql.NewSQLGenerator(gotosql.SQLDialectMySQL, false, map[string]string{"custompkg.myType": "INTEGER"})
	if err != nil {
		panic(err)
	}

	sql, err := sqlGenerator.Generate(ExampleObject{}, true)
	if err != nil {
		panic(err)
	}

	fmt.Println(sql)
}

type ExampleObject struct {
	Id          int64     `json:"id" dbType:"BIGINT AUTO_INCREMENT PRIMARY KEY"`
	Created     time.Time `json:"created"`
	Email       string    `json:"email" db:"emailAddress" dbType:"VARCHAR(20)"`
	Forename    string    `json:"forename"`
	Locked      bool      `json:"locked"`
	PhoneNumber string    `json:"phoneNumber"`
	SubTotal    float64   `json:"subTotal"`
	Surname     string    `json:"surname"`
}
```

The above code outputs this:

```sql
CREATE TABLE IF NOT EXISTS exampleObjects
(
    id
    INTEGER
    AUTO_INCREMENT
    PRIMARY
    KEY,
    created
    DATETIME,
    emailAddress
    VARCHAR
(
    20
) NOT NULL DEFAULT '',
    forename VARCHAR
(
    255
) NOT NULL DEFAULT '',
    locked TINYINT
(
    1
) NOT NULL DEFAULT 0,
    phoneNumber VARCHAR
(
    255
) NOT NULL DEFAULT '',
    subTotal DOUBLE NOT NULL DEFAULT 0.0,
    surname VARCHAR
(
    255
) NOT NULL DEFAULT ''
    );

CREATE TABLE IF NOT EXISTS exampleObjects_history
(
    id
    INTEGER,
    created
    DATETIME,
    emailAddress
    VARCHAR
(
    20
) NOT NULL DEFAULT '',
    forename VARCHAR
(
    255
) NOT NULL DEFAULT '',
    locked TINYINT
(
    1
) NOT NULL DEFAULT 0,
    phoneNumber VARCHAR
(
    255
) NOT NULL DEFAULT '',
    subTotal DOUBLE NOT NULL DEFAULT 0.0,
    surname VARCHAR
(
    255
) NOT NULL DEFAULT ''
    );

DROP TRIGGER IF EXISTS exampleObjects_audit;

DROP TRIGGER IF EXISTS exampleObjects_audit_first;

CREATE TRIGGER exampleObjects_audit
    BEFORE UPDATE
    ON exampleObjects
    FOR EACH ROW
BEGIN
    INSERT INTO exampleObjects_history (id, created, emailAddress, forename, locked, phoneNumber, subTotal, surname)
    VALUES (new.id, new.created, new.emailAddress, new.forename, new.locked, new.phoneNumber, new.subTotal,
            new.surname);
END;

CREATE TRIGGER exampleObjects_audit_first
    AFTER INSERT
    ON exampleObjects
    FOR EACH ROW
BEGIN
    INSERT INTO exampleObjects_history (id, created, emailAddress, forename, locked, phoneNumber, subTotal, surname)
    VALUES (new.id, new.created, new.emailAddress, new.forename, new.locked, new.phoneNumber, new.subTotal,
            new.surname);
END;
```

### Field Name

If the object's field name doesn't have a `db` tag, `gotosql` will convert the field's name to camel case and use it instead

### Override Data Type Assignment

When `gotosql` iterates over an object, it assigns the SQL data type based on this workflow (in this order):

1. If the field has a `dbType` tag, use it as the SQL data type
2. If the user provided a overarching data type override when initializing the generator, use this instead
3. Use the preconfigured SQL data types

_Note: If you want to specify a primary key or auto_increment field, use the `dbType` field tag_