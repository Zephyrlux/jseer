package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// AuditLog stores GM operations.
type AuditLog struct {
	ent.Schema
}

func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("operator").Default(""),
		field.String("action").Default(""),
		field.String("resource").Default(""),
		field.String("resource_id").Default(""),
		field.String("detail").Default(""),
		field.Time("created_at").Default(time.Now),
	}
}
