package db

import (
	"CheckUrls/pkg/logging"
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"time"
)

var (
	ErrNothingDone = fmt.Errorf("sql query did nothing")
	ConnManager    = &ConnectionManager{}
)

type DbConfig interface {
	GetDbHost() string
	GetDbPort() string
	GetDbUser() string
	GetDbPassword() string
	GetDbName() string
	GetDbSslmode() string
}

type ConnectionManager struct {
	conn *sql.DB
	log  *logging.Loggers
}

type sqlInfo struct {
	host     string
	port     string
	user     string
	password string
	dbname   string
	sslmode  string
}

func (c *ConnectionManager) Connect(cfg DbConfig) error {
	c.log = logging.NewLoggers("db", "connect")
	sq := sqlInfo{
		host:     cfg.GetDbHost(),
		port:     cfg.GetDbPort(),
		user:     cfg.GetDbUser(),
		password: cfg.GetDbPassword(),
		dbname:   cfg.GetDbName(),
		sslmode:  cfg.GetDbSslmode(),
	}

	connector := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		sq.host, sq.port, sq.user, sq.password, sq.dbname, sq.sslmode,
	)

	var err error
	c.conn, err = sql.Open("pgx", connector)
	if err != nil {
		c.log.ErrorLog().Str("when", "open connection").Err(err).Msg("failed to open connection")
		return err
	}
	if err := c.conn.Ping(); err != nil {
		c.log.ErrorLog().Str("when", "ping connection").Err(err).Msg("failed to ping connection")
		return err
	}

	return nil
}

func (c *ConnectionManager) Close() error {
	c.log = logging.NewLoggers("db", "close")
	if err := c.conn.Close(); err != nil {
		c.log.ErrorLog().Str("when", "clode connection").Err(err).Msg("failed to close connection")
		return err
	}
	return nil
}

func (c *ConnectionManager) Exec(query string, args ...interface{}) error {
	c.log = logging.NewLoggers("db", "exec")
	if err := c.conn.Ping(); err != nil {
		c.log.ErrorLog().Str("when", "ping connection").Err(err).Msg("failed to ping connection")
		return err
	}
	ctx := context.TODO()
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := c.conn.ExecContext(queryCtx, query, args...)
	if err != nil {
		c.log.ErrorLog().Str("when", "exec").Err(err).Msg("error at exec")
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		c.log.ErrorLog().Str("when", "get rows").Err(err).Msg("failed to get rows")
		return err
	}
	if rows == 0 {
		c.log.WarnLog().Msg("no rows")
		return ErrNothingDone
	}
	return nil
}

func (c *ConnectionManager) QueryRow(query string, args ...interface{}) (*sql.Row, func(), error) {
	c.log = logging.NewLoggers("db", "queryRow")
	if err := c.conn.Ping(); err != nil {
		c.log.ErrorLog().Str("when", "ping connection").Err(err).Msg("failed to ping connection")
		return nil, nil, err
	}
	ctx := context.TODO()
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	row := c.conn.QueryRowContext(queryCtx, query, args...)

	return row, cancel, nil
}

func (c *ConnectionManager) Query(query string, args ...interface{}) (*sql.Rows, func(), error) {
	c.log = logging.NewLoggers("db", "query")
	if err := c.conn.Ping(); err != nil {
		c.log.ErrorLog().Str("when", "ping connection").Err(err).Msg("failed to ping connection")
		return nil, nil, err
	}
	ctx := context.TODO()
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	rows, err := c.conn.QueryContext(queryCtx, query, args...)
	if err != nil {
		c.log.ErrorLog().Str("when", "get rows").Err(err).Msg("failed to get rows")
		defer cancel()
		return nil, nil, err
	}
	if rows == nil {
		c.log.WarnLog().Msg("no rows")
		return nil, cancel, ErrNothingDone
	}

	return rows, cancel, nil
}
