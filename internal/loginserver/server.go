package loginserver

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"jseer/internal/config"
	"jseer/internal/protocol"
	"jseer/internal/storage"

	"go.uber.org/zap"
)

const policyRequest = "<policy-file-request/>\x00"
const policyResponse = "<?xml version=\"1.0\"?><!DOCTYPE cross-domain-policy><cross-domain-policy><allow-access-from domain=\"*\" to-ports=\"*\" /></cross-domain-policy>\x00"

type Handler func(*Context)

type Context struct {
	Server *Server
	Conn   net.Conn
	CmdID  int32
	UserID uint32
	SeqID  int32
	Body   []byte
}

func (c *Context) Context() context.Context {
	return context.Background()
}

// Server handles login TCP protocol.
type Server struct {
	cfg      config.LoginConfig
	gameCfg  config.GameConfig
	logger   *zap.Logger
	store    storage.Store
	handlers map[int32]Handler
	mu       sync.RWMutex
}

func New(cfg config.LoginConfig, gameCfg config.GameConfig, store storage.Store, logger *zap.Logger) *Server {
	return &Server{
		cfg:      cfg,
		gameCfg:  gameCfg,
		logger:   logger,
		store:    store,
		handlers: make(map[int32]Handler),
	}
}

func (s *Server) Register(cmd int32, h Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[cmd] = h
}

func (s *Server) Start(ctx context.Context) error {
	if s.cfg.PolicyEnabled {
		go s.startPolicyServer(ctx)
	}
	ln, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return err
	}
	s.logger.Info("login server listening", zap.String("addr", s.cfg.Address))
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			s.logger.Warn("accept error", zap.Error(err))
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) startPolicyServer(ctx context.Context) {
	addr := net.JoinHostPort("0.0.0.0", fmtInt(s.cfg.PolicyPort))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Warn("policy server start failed", zap.Error(err))
		return
	}
	s.logger.Info("policy server listening", zap.String("addr", addr))
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			continue
		}
		_, _ = conn.Write([]byte(policyResponse))
		_ = conn.Close()
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Time{})
	reader := bufio.NewReader(conn)

	first, err := reader.ReadByte()
	if err != nil {
		return
	}
	if first == '<' {
		line, _ := reader.ReadBytes(0)
		msg := string(append([]byte{first}, line...))
		if msg == policyRequest {
			_, _ = conn.Write([]byte(policyResponse))
			return
		}
		// unknown text request, ignore
		return
	}

	lenBuf := make([]byte, 4)
	lenBuf[0] = first
	if _, err := io.ReadFull(reader, lenBuf[1:]); err != nil {
		return
	}

	for {
		pktLen := int(binary.BigEndian.Uint32(lenBuf))
		if pktLen < protocol.HeaderLen || pktLen > 1<<20 {
			s.logger.Warn("invalid packet length", zap.Int("len", pktLen))
			return
		}
		payload := make([]byte, pktLen-4)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return
		}
		data := append(lenBuf, payload...)
		_, _, cmdID, userID, seqID, body, err := protocol.ParsePacket(data)
		if err != nil {
			s.logger.Warn("parse packet failed", zap.Error(err))
			return
		}
		ctx := &Context{Server: s, Conn: conn, CmdID: cmdID, UserID: userID, SeqID: seqID, Body: body}
		s.dispatch(ctx)

		if _, err := io.ReadFull(reader, lenBuf); err != nil {
			return
		}
	}
}

func (s *Server) dispatch(ctx *Context) {
	s.mu.RLock()
	h, ok := s.handlers[ctx.CmdID]
	s.mu.RUnlock()
	if ok {
		h(ctx)
		return
	}
}

// SendResponse sends login response with result code.
func (s *Server) SendResponse(conn net.Conn, cmdID int32, userID uint32, result int32, body []byte) {
	resp := protocol.BuildResponse(cmdID, userID, result, body)
	_, _ = conn.Write(resp)
}

func fmtInt(v int) string {
	return strconv.Itoa(v)
}
