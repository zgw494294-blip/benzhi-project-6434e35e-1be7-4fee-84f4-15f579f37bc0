package storage

const schemaVersion = 1

func (s *Store) SchemaVersion() int { return schemaVersion }
