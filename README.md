# Remotebuild 
Build your applications remotely. Currently only AUR building is supported.

# Requirements
- PostgresDB (recommended)
- Docker
- Good Hardware 

# Compile
```bash
make build
```

# Concept
* You have one build server/VM where the server app runs
* Use the [client]("https://github.com/JojiiOfficial/RemoteBuildClient") to create/control jobs
* Only one compilation can be run at the same time
* A job exists of two sub types of jobs: Build job and Upload Job.
* A job is run inside a docker container on the server

# Setup
* Install docker 
* Install PostgreSql (recommended)

```bash
./main config create # Create an empty config
```

Fill out the `database` section. For help refer to [here](#database)

# Database
You can use PostgreSQL or Sqlite. Sqlite should only be used for debugging/testing purposes.<br>
Example of a recommended database setup:

```yaml
[...]
database:
  databasetype: postgres
  host: localhost
  username: "rbuild"
  database: "rbuild"
  pass: "mysecretpassword"
  databaseport: 5432
  sslmode: require
[...]
```

If you just want to test it and don't have a running PostgreSQL server, you can use following database config:

```yaml
[...]
  database:
    databasetype: sqlite
    databesFile: "testdb.db"
    host: localhost
    username: "postgres"
    database: "postgres"
    pass: "mysecretpassword"
    databaseport: 5432
    sslmode: disable
[...]
```
