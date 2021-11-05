# Commands

This is a list of supported Redis commands and how they are implemented.

The statuses are one of these:

- `stable` - it works as expected, can be used in production
- `limited-stable` - only parts of the Redis implementation are available, but they work as expected and can be used in production
- `beta` - it has been tested, but use with caution
- `alpha` - experimental, only for developers

| Command |      Status      | Comment                                                                                                                                             |
|:-------:|:----------------:|:----------------------------------------------------------------------------------------------------------------------------------------------------|
|  `DEL`  |     `stable`     | `DELETE FROM ... WHERE key = ...`                                                                                                                   |
|  `GET`  |     `stable`     | `SELECT value FROM ... WHERE key = ...`                                                                                                             |
| `KEYS`  | `limited-stable` | although, only `*`-globs are supported with `ILIKE` and `%`, because it is the most common scenario for me (to select keys starting with something) |
| `MGET`  |     `alpha`      | `SELECT value FROM ... WHERE key in (...)`, it might be slow for thousands of keys                                                                  |
| `PING`  |     `stable`     | `SELECT 1`                                                                                                                                          |
| `QUIT`  |     `stable`     | N/A (provided by `redcon`)                                                                                                                          |
|  `SET`  |     `stable`     | `INSERT` or `UPDATE`, i.e. ["UPSERT" method](https://www.postgresql.org/docs/14/sql-insert.html)                                                    |
