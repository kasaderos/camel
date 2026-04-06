package kvdb

type DB struct {
}

func New() *DB {
	return &DB{}
}

func (db *DB) Set(key string, value []byte) error {
	return nil
}

func (db *DB) Get(key string, value []byte) error {
	return nil
}
