# Migration Tool

This tool is used to create and run migrations in the database.

## Run Migration

```go
go run migration.go ./sql "host=localhost port=5432 user=root dbname=db_users password=fatannajuda sslmode=disable" up
```

## Down Migration

```go
go run migration.go ./sql "host=localhost port=5432 user=localhost dbname=db_users password=postgres sslmode=disable" down
```

## Create new SQL

```go
go run migration.go ./sql "host=localhost port=5432 user=localhost dbname=db_users sslmode=disable" create add_user_table sql
```
