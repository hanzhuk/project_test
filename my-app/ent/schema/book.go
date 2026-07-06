package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Book struct {
	ent.Schema
}

func (Book) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty(),
		field.String("author").Optional(),
		field.Float("price").Default(0.0),
	}
}

func (Book) Edges() []ent.Edge {
	return nil
}