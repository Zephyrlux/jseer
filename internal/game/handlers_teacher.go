package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
)

func registerTeacherHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(3001, handleRequestAddTeacher(deps, state))
	s.Register(3002, handleAnswerAddTeacher(deps, state))
	s.Register(3003, handleRequestAddStudent())
	s.Register(3004, handleAnswerAddStudent(deps, state))
	s.Register(3005, handleDeleteTeacher(deps, state))
	s.Register(3006, handleDeleteStudent(deps, state))
	s.Register(3007, handleExperienceSharedComplete(state))
	s.Register(3008, handleTeacherRewardComplete())
	s.Register(3009, handleMyExperiencePondComplete(state))
	s.Register(3010, handleSevenNoLoginComplete())
	s.Register(3011, handleGetMyExperienceComplete(state))
}

func handleRequestAddTeacher(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		_ = reader.ReadUint32BE()
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 3001, ctx.UserID, buf.Bytes())
		savePlayer(deps, ctx.UserID, state.GetOrCreateUser(ctx.UserID))
	}
}

func handleAnswerAddTeacher(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		studentID := reader.ReadUint32BE()
		accept := reader.ReadUint32BE()
		if accept == 1 && studentID > 0 {
			user := state.GetOrCreateUser(ctx.UserID)
			found := false
			for _, id := range user.StudentIDs {
				if id == studentID {
					found = true
					break
				}
			}
			if !found {
				user.StudentIDs = append(user.StudentIDs, studentID)
				if user.StudentID == 0 {
					user.StudentID = studentID
				}
			}
			student := state.GetOrCreateUser(studentID)
			student.TeacherID = ctx.UserID
			savePlayer(deps, ctx.UserID, user)
			savePlayer(deps, studentID, student)
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, accept)
		ctx.Server.SendResponse(ctx.Conn, 3002, ctx.UserID, buf.Bytes())
	}
}

func handleRequestAddStudent() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 3003, ctx.UserID, buf.Bytes())
	}
}

func handleAnswerAddStudent(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		teacherID := reader.ReadUint32BE()
		accept := reader.ReadUint32BE()
		if accept == 1 && teacherID > 0 {
			user := state.GetOrCreateUser(ctx.UserID)
			user.TeacherID = teacherID
			teacher := state.GetOrCreateUser(teacherID)
			found := false
			for _, id := range teacher.StudentIDs {
				if id == ctx.UserID {
					found = true
					break
				}
			}
			if !found {
				teacher.StudentIDs = append(teacher.StudentIDs, ctx.UserID)
				if teacher.StudentID == 0 {
					teacher.StudentID = ctx.UserID
				}
			}
			savePlayer(deps, ctx.UserID, user)
			savePlayer(deps, teacherID, teacher)
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, accept)
		ctx.Server.SendResponse(ctx.Conn, 3004, ctx.UserID, buf.Bytes())
	}
}

func handleDeleteTeacher(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		teacherID := user.TeacherID
		if teacherID > 0 {
			teacher := state.GetOrCreateUser(teacherID)
			next := teacher.StudentIDs[:0]
			for _, id := range teacher.StudentIDs {
				if id != ctx.UserID {
					next = append(next, id)
				}
			}
			teacher.StudentIDs = next
			if teacher.StudentID == ctx.UserID {
				teacher.StudentID = 0
			}
			savePlayer(deps, teacherID, teacher)
		}
		user.TeacherID = 0
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 3005, ctx.UserID, buf.Bytes())
	}
}

func handleDeleteStudent(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		next := user.StudentIDs[:0]
		for _, id := range user.StudentIDs {
			if id != targetID {
				next = append(next, id)
			}
		}
		user.StudentIDs = next
		if user.StudentID == targetID {
			user.StudentID = 0
		}
		if targetID > 0 {
			student := state.GetOrCreateUser(targetID)
			if student.TeacherID == ctx.UserID {
				student.TeacherID = 0
			}
			savePlayer(deps, targetID, student)
		}
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 3006, ctx.UserID, buf.Bytes())
	}
}

func handleExperienceSharedComplete(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, user.ExpPool)
		ctx.Server.SendResponse(ctx.Conn, 3007, ctx.UserID, buf.Bytes())
	}
}

func handleTeacherRewardComplete() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 3008, ctx.UserID, buf.Bytes())
	}
}

func handleMyExperiencePondComplete(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, user.ExpPool)
		ctx.Server.SendResponse(ctx.Conn, 3009, ctx.UserID, buf.Bytes())
	}
}

func handleSevenNoLoginComplete() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 3010, ctx.UserID, buf.Bytes())
	}
}

func handleGetMyExperienceComplete(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, user.ExpPool)
		ctx.Server.SendResponse(ctx.Conn, 3011, ctx.UserID, buf.Bytes())
	}
}
