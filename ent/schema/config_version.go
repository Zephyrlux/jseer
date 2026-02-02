package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ConfigVersion stores history for configuration changes.
type ConfigVersion struct {
	ent.Schema
}

func (ConfigVersion) Fields() []ent.Field {
	return []ent.Field{
		field.String("key"),
		field.Int64("version"),
		field.Bytes("value"),
		field.String("checksum").Default(""),
		field.String("operator").Default(""),
		field.Time("created_at").Default(time.Now),
	}
}

func (ConfigVersion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entry", ConfigEntry.Type).Ref("versions").Unique(),
	}
}
