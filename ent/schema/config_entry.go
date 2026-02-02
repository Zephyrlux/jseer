package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ConfigEntry stores current active config value.
type ConfigEntry struct {
	ent.Schema
}

func (ConfigEntry) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").Unique(),
		field.Bytes("value"),
		field.Int64("version").Default(1),
		field.String("checksum").Default(""),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (ConfigEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("versions", ConfigVersion.Type),
	}
}
