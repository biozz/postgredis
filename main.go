package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tidwall/redcon"
)

var (
	addr  = flag.String("addr", ":6380", "listen address")
	dbURL = flag.String("db", "postgres://postgres:postgres@localhost:5432/postgredis", "database to connect to")
	table = flag.String("table", "postgredis", "a table with key/value columns")
)

func main() {
	flag.Parse()
	go log.Printf("started server at %s", *addr)
	poolConfig, err := pgxpool.ParseConfig(*dbURL)
	poolConfig.ConnConfig.Logger = &Logger{}
	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Unable to connect to the database: %v", err)
	}
	_, err = pool.Exec(context.Background(), fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (key TEXT UNIQUE, value TEXT)", *table))
	if err != nil {
		log.Fatalf("Unable to create table %s: %v", *table, err)
	}
	s := Server{pool: pool}
	err = redcon.ListenAndServe(*addr, s.Handler, s.Accept, s.Closed)
	if err != nil {
		log.Fatal(err)
	}
}

type Server struct {
	pool *pgxpool.Pool
}

func (s *Server) Handler(conn redcon.Conn, cmd redcon.Command) {
	switch strings.ToLower(string(cmd.Args[0])) {
	default:
		conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
	case "ping":
		_, err := s.pool.Exec(context.Background(), "SELCT 1")
		if err != nil {
			conn.WriteError(err.Error())
			return
		}
		conn.WriteString("PONG")
	case "quit":
		conn.WriteString("OK")
		conn.Close()
	case "set":
		if len(cmd.Args) != 3 {
			conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
			return
		}
		_, err := s.pool.Exec(
			context.Background(),
			fmt.Sprintf(
				"INSERT INTO %s VALUES ('%s', '%s') ON CONFLICT (key) DO UPDATE SET value = '%s'", *table, cmd.Args[1], cmd.Args[2], cmd.Args[2],
			),
		)
		if err != nil {
			fmt.Println(err)
			conn.WriteError(err.Error())
			return
		}
		conn.WriteString("OK")
	case "get":
		if len(cmd.Args) != 2 {
			conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
			return
		}
		var val string
		err := s.pool.QueryRow(context.Background(), fmt.Sprintf("SELECT value FROM %s WHERE key = '%s'", *table, cmd.Args[1])).Scan(&val)
		if err != nil {
			fmt.Println(err)
			conn.WriteNull()
		} else {
			conn.WriteBulk([]byte(val))
		}
	case "del":
		if len(cmd.Args) != 2 {
			conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
			return
		}
		res, err := s.pool.Exec(context.Background(), fmt.Sprintf("DELETE FROM %s WHERE key = '%s'", *table, cmd.Args[1]))
		if err != nil {
			conn.WriteError(err.Error())
			return
		}
		conn.WriteInt(int(res.RowsAffected()))
	case "keys":
		if len(cmd.Args) != 2 {
			conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
			return
		}
		ilike := strings.Replace(string(cmd.Args[1]), "*", "%", -1)
		var keyCount int
		rows, err := s.pool.Query(context.Background(), fmt.Sprintf("SELECT key FROM %s WHERE key ILIKE '%s'", *table, ilike))
		if err != nil {
			conn.WriteError(err.Error())
			return
		}
		keys := make([]string, 0)
		for rows.Next() {
			var key string
			err = rows.Scan(&key)
			if err != nil {
				conn.WriteError(err.Error())
				return
			}
			keyCount++
			keys = append(keys, key)
		}
		conn.WriteArray(keyCount)
		for _, key := range keys {
			conn.WriteAny(key)
		}
	}
}

func (s *Server) Accept(conn redcon.Conn) bool {
	// Use this function to accept or deny the connection.
	// log.Printf("accept: %s", conn.RemoteAddr())
	return true
}

func (s *Server) Closed(conn redcon.Conn, err error) {
	// This is called when the connection has been closed
	// log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
}

type Logger struct{}

func (l *Logger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	time, gotTime := data["time"]
	sql, gotSql := data["sql"]
	if gotTime && gotSql {
		log.Printf("%s %s\n", time, sql)
	}
}
