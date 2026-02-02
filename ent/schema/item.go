package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Item holds inventory item data.
type Item struct {
	ent.Schema
}

func (Item) Fields() []ent.Field {
	return []ent.Field{
		field.Int("player_id"),
		field.Int("item_id"),
		field.Int("count").Default(1),
		field.String("meta").Default(""),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Item) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("player", Player.Type).Ref("items").Field("player_id").Unique().Required(),
	}
}
