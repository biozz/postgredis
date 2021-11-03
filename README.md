# postgredis

A wild idea of having Redis-compatible server with PostgreSQL backend.

## Getting started

As a binary:

```bash
./postgredis -addr=:6380 -db=postgres://postgres:postgres@localhost:5432/postgredis -table=postgredis
```

As a docker container:

```bash
docker run -it postgredis:latest -addr=:6380 -db=postgres://postgres:postgres@db:5432/postgredis -table=postgredis
```

Note: it was intended to run inside docker network with postgres in a container on the same network, see `docker-compose.yml`

You can use redis-cli and play around with supported commands (see below):

```bash
> redis-cli -p 6380
127.0.0.1:6380> set x 123
OK
127.0.0.1:6380> set x 1
OK
127.0.0.1:6380> get x
"1"
```

or connect with various redis-compatible libraries (see `example/example.go`).

## Why

I like how Redis interface works. It is easy to understand and you can bootstrap new projects quickly without dealing with persistant storage.
I also don't have to deal with migrations management and sql-related libraries. But Redis itself is not good as a [persistent](https://redis.io/topics/persistence) storage.

I like PostgreSQL, it is my goto database. I have a couple of web services using it at the moment. And I try to stick with PostgreSQL when any of the self-hosted appliances provide it as a storage choice.

I've been using [bitcask](https://git.mills.io/prologic/bitcask) a lot recently, and I liked how you can start bitcask server and use libraries like [redigo](https://github.com/gomodule/redigo) to interact with the server. I then found out that writing Redis interfaces for any storage backend is quite easy with [redcon](https://github.com/tidwall/redcon) library. I even looked at solutions like [LedisDB](https://ledisdb.io/), but it seemed unmaintained and too bloated for my use cases.

So I decided to implement PostgreSQL backend for Redis' interface.

## How it works

When you first start postgredis, it will create a table with two columns: unique `key` and `value`. They both are in [text](https://www.postgresql.org/docs/14/datatype-character.html) format. Unique `key` helps with concise "upsert" queries. Thanks to [pgx](https://github.com/jackc/pgx) I don't have think about how to maintain a connection, I just start a [pool](https://github.com/jackc/pgx/tree/master/pgxpool). There is also a logging of sql queries, which is good for debugging.

These Redis commands are supported:

- `SET` - `INSERT` or `UPDATE`, i.e. ["UPSERT" method](https://www.postgresql.org/docs/14/sql-insert.html)
- `GET` - `SELECT value FROM ... WHERE key = ...`
- `DEL` - `DELETE FROM ... WHERE key = ...`
- `KEYS` - although, only `*`-globs are supported with `ILIKE` and `%`, because it is the most common scenario for me (to select keys starting with something)
- `PING` - `SELECT 1`
- `QUIT`

## Future plans

- make use of query args from the redis connection uri to specify table - http://www.iana.org/assignments/uri-schemes/prov/redis
- implement pub/sub mechanizms
- implement other data types, ex. sets
- implement scan
- add ttl handling
- indexing and other query optimizations
