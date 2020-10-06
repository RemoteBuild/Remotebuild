# Remotebuild
Compile applications, packages, repositories remotely.

# Supported build types
- [x] AUR packages

# Requirements
- PostgresSql database (recommended)
- Docker

# Compile
```bash
make build
```

# Concept
* You have one build server/VM where the server app needs to be up and running
* Use the [client]("https://github.com/JojiiOfficial/RemoteBuildClient") to create/control jobs
* A job exists of two sub types of jobs: Build job and Upload Job
* Only one job can be run at the same time

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
<br>

If you just want to test it and don't have a running PostgreSQL server, you can use following database config:

```yaml
[...]
database:
  databasetype: sqlite
  databesFile: "testdb.db"
[...]
```
