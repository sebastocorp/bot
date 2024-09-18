package database

import (
	"bot/api/v1alpha1"
	"bot/internal/logger"
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

type ManagerT struct {
	Ctx       context.Context
	Connector driver.Connector
	// MySQL MySQLT
	Table string
}

// type MySQLT struct {
// 	MySQLCredsT
// 	Connector driver.Connector
// }

// type MySQLCredsT struct {
// 	Host     string
// 	Port     string
// 	User     string
// 	Pass     string
// 	Database string
// 	Table    string
// }

type QueryObjectResultT struct {
	Id         *int
	BlobPath   *string
	BucketName *string
	Md5Sum     *string
	CreatedAt  *string
	UpdatedAt  *string
}

type ObjectT struct {
	Bucket string
	Path   string
	MD5    string
}

func NewManager(ctx context.Context, db v1alpha1.DatabaseT) (man ManagerT, err error) {
	man.Ctx = ctx
	// man.MySQL.Host = db.Host
	// man.MySQL.Port = db.Port
	// man.MySQL.User = db.Username
	// man.MySQL.Pass = db.Password
	// man.MySQL.Database = db.Database
	// man.MySQL.Table = db.Table

	if db.Table == "" {
		err = fmt.Errorf("database table not provided")
		return man, err
	}

	// Get a database handle.
	man.Connector, err = mysql.NewConnector(&mysql.Config{
		User:                 db.Username,
		Passwd:               db.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", db.Host, db.Port),
		DBName:               db.Database,
		AllowNativePasswords: true,
	})

	return man, err
}

// GetObject TODO
func (m *ManagerT) GetObject(object v1alpha1.ObjectT) (result QueryObjectResultT, occurrences int, err error) {

	// Get a database handle.
	db := sql.OpenDB(m.Connector)
	defer db.Close()

	queryClause := fmt.Sprintf("SELECT * FROM %s WHERE bucket_name='%s' AND blob_path='%s';",
		m.Table, object.BucketName, object.ObjectPath)

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
func (m *ManagerT) InsertObject(object ObjectT) (err error) {

	// Get a database handle.
	db := sql.OpenDB(m.Connector)
	defer db.Close()

	// Insert the object into the database.
	queryClause := fmt.Sprintf("INSERT INTO %s (blob_path,md5sum,bucket_name) VALUES ('%s', '%s', '%s');",
		m.Table, object.Path, object.MD5, object.Bucket)

	rows, err := db.Query(queryClause)
	if err != nil {
		return err
	}
	rows.Close()

	return err
}

func (m *ManagerT) InsertObjectsIfNotExist(objectList []v1alpha1.DatabaseRequestT) (err error) {

	// Get a database handle.
	db := sql.OpenDB(m.Connector)
	defer db.Close()

	for _, object := range objectList {
		var exists bool
		searchQueryClause := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE bucket_name = ? AND blob_path = ?)", m.Table)
		err = db.QueryRow(searchQueryClause, object.BucketName, object.ObjectPath).Scan(&exists)
		if err != nil {
			logger.Logger.Errorf("unable to check object {bucket: '%s', path: '%s'}: %s", object.BucketName, object.ObjectPath, err.Error())
			continue
		}

		if !exists {
			// Insert the object into the database.
			insertQueryClause := fmt.Sprintf("INSERT INTO %s (blob_path,md5sum,bucket_name) VALUES (?, ?, ?)", m.Table)
			_, err := db.Exec(insertQueryClause, object.ObjectPath, object.MD5, object.BucketName)
			if err != nil {
				logger.Logger.Errorf("unable to insert object {bucket: '%s', path: '%s'}: %s", object.BucketName, object.ObjectPath, err.Error())
				continue
			}
		}
	}

	return err
}
