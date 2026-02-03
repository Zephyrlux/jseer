package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Pet represents a captured creature.
type Pet struct {
	ent.Schema
}

func (Pet) Fields() []ent.Field {
	return []ent.Field{
		field.Int("player_id"),
		field.Int("species_id"),
		field.Int("level").Default(1),
		field.Int("exp").Default(0),
		field.Int("hp").Default(0),
		field.Int64("catch_time").Default(0),
		field.Int("dv").Default(31),
		field.String("nature").Default("normal"),
		field.String("skills").Default(""),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Pet) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("player", Player.Type).Ref("pets").Field("player_id").Unique().Required(),
	}
}
