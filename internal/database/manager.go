package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"bot/internal/objectStorage"

	"github.com/go-sql-driver/mysql"
)

type ManagerT struct {
	Ctx   context.Context
	MySQL MySQLT
}

type MySQLT struct {
	MySQLCredsT
	Connector driver.Connector
}

type MySQLCredsT struct {
	Host     string
	Port     string
	User     string
	Pass     string
	Database string
	Table    string
}

type QueryObjectResultT struct {
	Id         *int
	BlobPath   *string
	BucketName *string
	Md5Sum     *string
	CreatedAt  *string
	UpdatedAt  *string
}

func NewManager(ctx context.Context, mysqlCreds MySQLCredsT) (man ManagerT, err error) {
	man.Ctx = ctx
	man.MySQL.Host = mysqlCreds.Host
	man.MySQL.Port = mysqlCreds.Port
	man.MySQL.User = mysqlCreds.User
	man.MySQL.Pass = mysqlCreds.Pass
	man.MySQL.Database = mysqlCreds.Database
	man.MySQL.Table = mysqlCreds.Table

	if man.MySQL.Table == "" {
		err = fmt.Errorf("database table not provided")
		return man, err
	}

	// Get a database handle.
	man.MySQL.Connector, err = mysql.NewConnector(&mysql.Config{
		User:                 man.MySQL.User,
		Passwd:               man.MySQL.Pass,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", man.MySQL.Host, man.MySQL.Port),
		DBName:               man.MySQL.Database,
		AllowNativePasswords: true,
	})

	return man, err
}

// GetObject TODO
func (m *ManagerT) GetObject(object objectStorage.ObjectT) (result QueryObjectResultT, occurrences int, err error) {

	// Get a database handle.
	db := sql.OpenDB(m.MySQL.Connector)
	defer db.Close()

	queryClause := fmt.Sprintf("SELECT * FROM %s WHERE bucket_name='%s' AND blob_path='%s';",
		m.MySQL.Table, object.BucketName, object.ObjectPath)

	rows, err := db.Query(queryClause)
	if err != nil {
		return result, occurrences, err
	}
	defer rows.Close()

	//
	result = QueryObjectResultT{
		Id:         new(int),
		BlobPath:   new(string),
		BucketName: new(string),
		Md5Sum:     new(string),
		CreatedAt:  new(string),
		UpdatedAt:  new(string),
	}

	for rows.Next() {
		err = rows.Scan(result.Id, result.BlobPath, result.Md5Sum, result.BucketName, result.CreatedAt, result.UpdatedAt)
		occurrences++
	}

	return result, occurrences, err
}

// InsertObject TODO
func (m *ManagerT) InsertObject(object objectStorage.ObjectT) (err error) {

	// Get a database handle.
	db := sql.OpenDB(m.MySQL.Connector)
	defer db.Close()

	// Insert the object into the database.
	queryClause := fmt.Sprintf("INSERT INTO %s (blob_path,md5sum,bucket_name) VALUES ('%s', '%s', '%s');",
		m.MySQL.Table, object.ObjectPath, object.Etag, object.BucketName)

	rows, err := db.Query(queryClause)
	if err != nil {
		return err
	}
	rows.Close()

	return err
}
