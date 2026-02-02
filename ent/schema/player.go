package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Player holds core character data.
type Player struct {
	ent.Schema
}

func (Player) Fields() []ent.Field {
	return []ent.Field{
		field.Int("account_id"),
		field.String("nick").NotEmpty(),
		field.Int("level").Default(1),
		field.Int64("coins").Default(0),
		field.Int64("gold").Default(0),
		field.Int("map_id").Default(1),
		field.Int("pos_x").Default(300),
		field.Int("pos_y").Default(300),
		field.Time("last_login_at").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Player) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", Account.Type).Ref("players").Field("account_id").Unique().Required(),
		edge.To("pets", Pet.Type),
		edge.To("items", Item.Type),
	}
}
