package mysql

// ConnectionConfig represents the configuration settings for a MySQL database.
// It provides details about the database's connection parameters as well as
// various connection pool settings that can be used to optimize the database
// connections for a specific use case or environment.
//
// Fields:
//   - Host: The host name or IP address of the MySQL server.
//   - Port: The port number on which the MySQL server is listening.
//   - UserName: The username to use when connecting to the MySQL database.
//   - Password: The password to use when connecting to the MySQL database.
//   - DbName: The name of the specific MySQL database to connect to.
//   - MaxOpenConn: The maximum number of open connections to the database. This
//     can be used to control the size of the connection pool.
//   - MaxIdleConn: The maximum number of idle connections that can be maintained
//     in the connection pool.
//   - MaxConnLifetime: The maximum duration in seconds a connection can remain open.
//     After this duration, the connection will be closed and removed from the pool.
//   - ConnMaxIdleTime: The maximum duration in seconds a connection can remain idle
//     before it's closed and removed from the pool.
type ConnectionConfig struct {
	Host            string
	Port            string
	UserName        string
	Password        string
	DbName          string
	MaxOpenConn     int64
	MaxIdleConn     int64
	MaxConnLifetime int64
	ConnMaxIdleTime int64
}

type ReadQueryInput struct {
	Query             string
	Params            []interface{}
	UseCache          bool
	CapitaliseColumns bool
}
