# Chill && Code #1 - In-Memory DB in Go in (15m)

## `DB Features`

- set key
- get key
- delete key
- manage client connections
- warn clients before closing
- backup memory records before exiting

## `How to Test`

#### `server`
```bash
go run main.go
```

#### `client`
```bash
telnet localhost 8080

# set a key in memory db
set key val

# get a key from the memory db
get key

# delete a key from the memory db
delete key
```

#### `db.json`

Check out the db.json for saved data
