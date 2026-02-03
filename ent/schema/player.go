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
		field.Int("map_type").Default(0),
		field.Int("pos_x").Default(300),
		field.Int("pos_y").Default(300),
		field.Int("last_map_id").Default(1),
		field.Int64("color").Default(0x66CCFF),
		field.Int64("texture").Default(1),
		field.Int64("energy").Default(100),
		field.Int64("fight_badge").Default(0),
		field.Int64("time_today").Default(0),
		field.Int64("time_limit").Default(86400),
		field.Int64("teacher_id").Default(0),
		field.Int64("student_id").Default(0),
		field.Int64("cur_title").Default(0),
		field.String("task_status").Default("{}"),
		field.String("task_bufs").Default("{}"),
		field.String("friends").Default("[]"),
		field.String("blacklist").Default("[]"),
		field.String("achievements").Default("[]"),
		field.String("titles").Default("[]"),
		field.String("team_info").Default("{}"),
		field.String("student_ids").Default("[]"),
		field.Int64("room_id").Default(0),
		field.String("fitments").Default("[]"),
		field.String("nono_info").Default("{}"),
		field.String("mailbox").Default("[]"),
		field.Int64("current_pet_id").Default(0),
		field.Int64("current_pet_catch_time").Default(0),
		field.Int64("current_pet_dv").Default(31),
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
