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
	sqlGenerator, err := gotosql.NewSqlGenerator(gotosql.SqlDialectMySql, false, map[string]string{"custompkg.myType": "INTEGER"})
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
	Id          int64     `json:"id"`
	Created     time.Time `json:"created"`
	Email       string    `json:"email" gotosql:"VARCHAR(20)"`
	Forename    string    `json:"forename"`
	Locked      bool      `json:"locked"`
	PhoneNumber string    `json:"phoneNumber"`
	SubTotal    float64   `json:"subTotal"`
	Surname     string    `json:"surname"`
}
```

The above code outputs this:
```sql
CREATE TABLE IF NOT EXISTS exampleobjects (
    id BIGINT NOT NULL DEFAULT 0,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    email VARCHAR(20) NOT NULL DEFAULT '',
    forename VARCHAR(255) NOT NULL DEFAULT '',
    locked TINYINT(1) NOT NULL DEFAULT 0,
    phoneNumber VARCHAR(255) NOT NULL DEFAULT '',
    subTotal DOUBLE NOT NULL DEFAULT 0.0,
    surname VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS exampleobjects_history (
    id BIGINT NOT NULL DEFAULT 0,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    email VARCHAR(20) NOT NULL DEFAULT '',
    forename VARCHAR(255) NOT NULL DEFAULT '',
    locked TINYINT(1) NOT NULL DEFAULT 0,
    phoneNumber VARCHAR(255) NOT NULL DEFAULT '',
    subTotal DOUBLE NOT NULL DEFAULT 0.0,
    surname VARCHAR(255) NOT NULL DEFAULT ''
);

DROP TRIGGER IF EXISTS exampleobjects_audit;

DROP TRIGGER IF EXISTS exampleobjects_audit_first;

CREATE TRIGGER exampleobjects_audit BEFORE UPDATE ON exampleobjects FOR EACH ROW BEGIN
    INSERT INTO exampleobjects_history (
        id, created, email, forename, locked, phoneNumber, subTotal, surname
    ) VALUES (
        new.id, new.created, new.email, new.forename, new.locked, new.phoneNumber, new.subTotal, new.surname
    );
END;

CREATE TRIGGER exampleobjects_audit_first AFTER INSERT ON exampleobjects FOR EACH ROW BEGIN
    INSERT INTO exampleobjects_history (
        id, created, email, forename, locked, phoneNumber, subTotal, surname
    ) VALUES (
        new.id, new.created, new.email, new.forename, new.locked, new.phoneNumber, new.subTotal, new.surname
    );
END;
```

### Override Data Type Assignment
When `gotosql` iterates over an object, it assigns the SQL data type based on this workflow (in this order):
1. If the field has a `gotosql` tag, use it as the SQL data type
2. If the user provided a overarching data type override when initializing the generator, use this instead
3. Use the preconfigured SQL data typess
