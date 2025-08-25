# sqljson
A Go utility library providing an ergonomic solution to struct-to-JSON marshalling for sql.

Each DB driver has its own set of quirks when working with the JSON typed columns.  
This library provides a simple and ergonomic way to marshal/scan any field type into JSON and back, regardless of the DB driver you're using.

Sponsored by `stdlib` mode of `pgx`, `pgbouncer` weirdnesses, and indecisive `lib/pq` scanner :)

## Installation
```bash
go get github.com/utrack/sqljson@latest
```

## Usage
```go
import (
    "github.com/utrack/sqljson"
    "database/sql"
    "context"
)

type User struct {
    ID    int    `json:"id"` // DB column name: id, type: int
    Name  string `json:"name"` // DB column name: name, type: text
    Email string `json:"email"` // DB column name: email, type: text
    Tags []string `json:"tags"` // DB column name: tags, type: jsonb
}

func selectRows() error {
    rows, err := db.Query("SELECT id,name,email,tags FROM users")
    if err != nil {
        return err
    }
    defer rows.Close()
    
    // instead of doing this:
    for rows.Next() {
        var user User
        var tagsBuf json.RawMessage
        if err := rows.Scan(&user.ID, &user.Name, &user.Email, &tagsBuf); err != nil {
            return err
        }
        if err := json.Unmarshal(tagsBuf, &user.Tags); err != nil {
            return err
        }
    }

    // do that!
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Name, &user.Email, sqljson.As(&user.Tags)); err != nil {
            return err
        }
    }
}

func insertRow(ctx context.Context,u User) error {
   q := "INSERT INTO users (id,name,email,tags) VALUES ($1,$2,$3,$4)"
   
   // works for marshalling, too!
   _,err := db.ExecContext(ctx,q, u.ID, u.Name, u.Email, sqljson.As(u.Tags))
   
   return err
}

// this example describes usage with sqlx, *cough* gorm and other ORMs
// that can scan the columns directly into a struct
type intermediateUser struct {
	ID    int                     `json:"id" db:"id"`
	Name  string                  `json:"name" db:"name"`
	Email string                  `json:"email" db:"email"`
	Tags  sqljson.Field[[]string] `json:"tags" db:"tags"`
}

func sqlxSelect(ctx context.Context) error {
   q := "SELECT id,name,email,tags FROM users WHERE id = 1"
   
   // here, sqlx will scan tags directly into intermediateUser.Tags...
   var user intermediateUser
   if err := sqlx.GetContext(ctx,db,q,&user); err != nil {
      return err
   }
   
   // which becomes an unmarshalled []string
   ret := User{
       ID: user.ID,
       Name: user.Name,
       Email: user.Email,
       Tags: user.Tags.Get(),
   }
   
   return nil
}
```
