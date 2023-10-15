package mysql

import "fmt"

type DSN struct {
	User     string
	Password string
	Host     string
	Database string
}

func (dsn *DSN) String() string {
	return fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true", dsn.User, dsn.Password, dsn.Host, dsn.Database)
}
