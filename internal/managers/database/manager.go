package database

import (
	"bot/api/v1alpha3"
	"bot/internal/pools"
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

type ManagerT struct {
	Ctx       context.Context
	Connector driver.Connector
}

// type QueryObjectResultT struct {
// 	Id         *int
// 	BlobPath   *string
// 	BucketName *string
// 	Md5Sum     *string
// 	CreatedAt  *string
// 	UpdatedAt  *string
// }

type ObjectT struct {
	Bucket string
	Path   string
	MD5    string
}

func NewManager(ctx context.Context, db v1alpha3.DatabaseT) (man ManagerT, err error) {
	man.Ctx = ctx

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

func (m *ManagerT) InsertObjectListIfNotExist(table string, objectList []pools.DatabaseRequestT) (err error) {
	objectListLen := len(objectList)

	insertQueryClause := fmt.Sprintf("INSERT IGNORE INTO %s (blob_path,md5sum,bucket_name) VALUES ",
		table,
	)
	for index, object := range objectList {
		// Insert the object into the database.
		insertQueryClause += fmt.Sprintf("('%s', '%s', '%s')",
			object.ObjectPath,
			object.MD5,
			object.BucketName,
		)
		if index < objectListLen-1 {
			insertQueryClause += ", "
		}
	}
	insertQueryClause += ";"

	// Get a database handle.
	db := sql.OpenDB(m.Connector)
	defer db.Close()

	_, err = db.Exec(insertQueryClause)

	return err
}
