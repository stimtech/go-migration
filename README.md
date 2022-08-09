# Stim Go Migration lib #
Library for database sql migrations. Update your databases using incremental SQL-scripts.

## Install ##
``` bash
go get github.com/stimtech/go-migration
```

## Usage ##
Add your SQL files to the `db/migrations` directory.

Create the migration service using an (in this example, sqlite) `*sql.DB`
``` go
db, err := sql.Open("sqlite3", "db.sqlite")
logger, err := zap.NewProduction()
m := migration.New(db, logger) // use zap.NewNop() if you don't want logs
```
Then start the migration
``` go
err = m.Migrate()
```

## Databases ##
This library is tested with `SQLite`, `MySQL` and `PostgreSQL`, but will probably work with many other SQL databases.

## How it works ##
Running `Migration()` will do the following things:

- Create the `migration` and `migration_lock` tables if they don't exist already.

- Inserts value in `migration_lock`.

  - If the insert fails (another process has the lock), it will try again every 5 seconds for a minute. If it still doesn't have the lock it will return an error.
  - The lock value is automatically removed after 15 minutes, or when the migration finishes.

- All previously applied migrations are fetched from the `migration` table.

- Lists all SQL-files in the `db/migrations` folder.

- For each file, in alphabetical order:

  - If the file has not been applied before, apply it now.

    - If the file cannot be applied, roll back the entire file (if possible), and return an error.
    - If apply is successful, add the filename and checksum to `migration`.

  - If the file has been applied before, compare the file's checksum with the checksum in `migration`. Return an error if they differ.

Note that some databases, MySQL for example, can not roll back DDL altering statements (like `CREATE` or `MODIFY`)

## Configuration ##
These are settings that can be configured.

- `TableName`: the table where all applied migrations are stored. Defaults to `migration`

- `LocKTableName`: the table where the lock is held. Defaults to `migration_lock`

- `Folder`: the folder where all migration SQL files are. Defaults to `db/migrations`

- `LockTimeoutMinutes`: how long a lock can be held before it times out, in minutes. Defaults to 15

## Design decisions and philosophy ##

### Checksums ###
For consistency between environments, the SQL files should never be updated once applied to a database (outside of development environment).
The checksums make sure that the files in `db/migrations` are identical to the ones applied to the database.

If changes are needed, a new SQL file with those changes should be created.

### Transactions ###
All changes in a single file are applied in a transaction. That way no partial migrations are ever present in the database.

### Lock ###
The locking mechanism allows several instances of the same application to be deployed at the same time.
Only one of them will apply the migrations, to avoid conflicts.
The other instances will wait until the first one completes its migration.

We use a table with a single primary key column to manage the locks.
This type of locking is supported by most SQL databases.

### Out-of-order versioning ###
It is not always known in which order features will be merged to trunk, when the work is started.
With out-of-order versioning, features can be merged in any order, without having to sync and rename migration files.

### No down-migrations ###
Down-migrations (also called roll-backs) are hard to test, and may change or destroy production data in unexpected ways.
When there is a problem, create a new up-migration that fixes the problem.

### Backwards compatibility ###
To minimize downtime, it is recommended that all database migrations are compatible with the previous version of the application.
This will allow reverting the application to a previous version without having to do down-migrations.

### Recommended naming convention ###
All files must have names that is later in the lexicographic order than previous files. 
It is recommended to start all file names with the date they are created, possibly followed by a ticket number, and a short description.
Like so:

- `2022-05-21-#2-initial-db.sql`

- `2022-05-28-#13-create-users-table.sql`

- `2022-06-01-#22-add-email-to-users.sql`
