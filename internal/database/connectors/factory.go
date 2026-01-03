package connectors

import (
	"fmt"
)

// ConnectorFactory creates database connectors based on database type
type ConnectorFactory struct{}

// NewConnectorFactory creates a new connector factory
func NewConnectorFactory() *ConnectorFactory {
	return &ConnectorFactory{}
}

// CreateConnector creates a connector for the specified database type
func (f *ConnectorFactory) CreateConnector(dbType DatabaseType) (Connector, error) {
	switch dbType {
	case DatabaseTypePostgreSQL:
		return NewPostgreSQLConnector(), nil
	case DatabaseTypeMySQL:
		return NewMySQLConnector(), nil
	case DatabaseTypeSQLite:
		return NewSQLiteConnector(), nil
	case DatabaseTypeMongoDB:
		return NewMongoDBConnector(), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDatabase, dbType)
	}
}

// ValidateDatabaseType validates if the database type is supported
func (f *ConnectorFactory) ValidateDatabaseType(dbType DatabaseType) bool {
	switch dbType {
	case DatabaseTypePostgreSQL, DatabaseTypeMySQL, DatabaseTypeSQLite, DatabaseTypeMongoDB:
		return true
	default:
		return false
	}
}

// GetSupportedDatabaseTypes returns a list of supported database types
func (f *ConnectorFactory) GetSupportedDatabaseTypes() []DatabaseType {
	return []DatabaseType{
		DatabaseTypePostgreSQL,
		DatabaseTypeMySQL,
		DatabaseTypeSQLite,
		DatabaseTypeMongoDB,
	}
}
