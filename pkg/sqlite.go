package pkg

import (
	"context"
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

// sqlitePragmas are applied to every new connection via the DSN, since most
// pragmas are per-connection. busy_timeout goes first so it's set before the WAL switch.
const sqlitePragmas = "_pragma=busy_timeout(5000)" +
	"&_pragma=journal_mode(WAL)" +
	"&_pragma=synchronous(NORMAL)" +
	"&_pragma=foreign_keys(ON)"

func ConnectToSQLite(ctx context.Context, path string) (writeDB, readDB *sql.DB, err error) {
	// _txlock=immediate takes the write lock at BEGIN, avoiding SQLITE_BUSY
	// when a tx upgrades from read to write
	writeDB, err = sql.Open("sqlite", "file:"+path+"?"+sqlitePragmas+"&_txlock=immediate")
	if err != nil {
		return nil, nil, err
	}
	writeDB.SetMaxOpenConns(1)
	writeDB.SetConnMaxIdleTime(time.Minute)

	readDB, err = sql.Open("sqlite", "file:"+path+"?"+sqlitePragmas)
	if err != nil {
		return nil, nil, err
	}
	readDB.SetMaxOpenConns(100)
	readDB.SetConnMaxIdleTime(time.Minute)
	return
}
