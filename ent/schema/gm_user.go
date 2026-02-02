package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// GMUser represents a GM account.
type GMUser struct {
	ent.Schema
}

func (GMUser) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Unique(),
		field.String("password_hash").NotEmpty(),
		field.String("status").Default("active"),
		field.Time("last_login_at").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (GMUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roles", Role.Type).Ref("gm_users"),
	}
}
