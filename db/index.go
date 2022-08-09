package db

// DBCreator DBCreator
// nolint
type DBCreator interface {
	Exists(name string) (bool, error)
	Create(name string, table, definition string) (bool, error)
	Delete(name string) (bool, error)
	GetVersion() int32
}
