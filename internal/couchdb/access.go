package couchdb

import (
	"context"
	"couchdb-proxy/internal/pg"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

const (
	tokenExpirationDays = 30
)

const (
	getAuthUserSql = "SELECT user_id " +
		"FROM users_token tokens " +
		"WHERE tokens.key = $1 AND tokens.created > $2"
	checkProjectAccessSql = "SELECT count(*) " +
		"FROM projects_project projects " +
		"WHERE projects.db_name = $1 " +
		"  AND (projects.is_public " +
		"		OR projects.owner_id = $2 " +
		"		OR EXISTS (SELECT 1 " +
		"				FROM projects_projectmember members " +
		"				WHERE members.project_id = projects.id AND members.user_id = $2" +
		" 			) " +
		"       )"
)

func isAccessAllowed(database string, auth string) (allowed bool, err error) {
	dbpool := pg.GetConnectionPool()
	conn, err := dbpool.Acquire(context.Background())
	if err != nil {
		return
	}
	defer conn.Release()

	userId, err := retrieveUser(conn, auth)
	if err != nil {
		return
	}

	if userId == -1 {
		return
	}

	if database == "" {
		return true, nil
	}

	allowed, err = checkProjectAccess(conn, database, userId)

	return
}

func retrieveUser(conn *pgxpool.Conn, auth string) (user int, err error) {
	now := time.Now()
	expiration := now.AddDate(0, 0, -tokenExpirationDays)

	var userId *int
	err = conn.QueryRow(context.Background(), getAuthUserSql, auth, expiration).Scan(&userId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return -1, nil
		}
		return
	}
	return *userId, nil
}

func checkProjectAccess(conn *pgxpool.Conn, database string, user int) (allowed bool, err error) {
	var rowsCount *int
	err = conn.QueryRow(context.Background(), checkProjectAccessSql, database, user).Scan(&rowsCount)
	if err != nil {
		return
	}
	return *rowsCount > 0, nil
}
