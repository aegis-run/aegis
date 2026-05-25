package memory

import "github.com/hashicorp/go-memdb"

const (
	TableSchema = "schema"
)

const (
	IndexID        = "id"
	IndexCreatedAt = "created_at"
)

var Schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		TableSchema: {
			Name: TableSchema,
			Indexes: map[string]*memdb.IndexSchema{
				IndexID: {
					Name:    IndexID,
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Hash"},
				},

				IndexCreatedAt: {
					Name:    IndexCreatedAt,
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "CreatedAt"},
				},
			},
		},
	},
}
