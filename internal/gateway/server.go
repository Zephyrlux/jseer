package gateway

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"jseer/internal/config"
	"jseer/internal/protocol"

	"go.uber.org/zap"
)

type Handler func(*Context)

// Context carries request data for handlers.
type Context struct {
	Server *Server
	Conn   net.Conn
	CmdID  int32
	UserID uint32
	SeqID  int32
	Body   []byte
}

// Server handles the TCP gateway for AS3 clients.
type Server struct {
	cfg           config.GatewayConfig
	logger        *zap.Logger
	handlers      map[int32]Handler
	defaultHandle Handler
	mu            sync.RWMutex
}

func New(cfg config.GatewayConfig, logger *zap.Logger) *Server {
	return &Server{
		cfg:      cfg,
		logger:   logger,
		handlers: make(map[int32]Handler),
	}
}

func (s *Server) Register(cmd int32, h Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[cmd] = h
}

func (s *Server) RegisterIfAbsent(cmd int32, h Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.handlers[cmd]; ok {
		return
	}
	s.handlers[cmd] = h
}

func (s *Server) SetDefault(h Handler) {
	s.defaultHandle = h
}

func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return err
	}
	s.logger.Info("gateway listening", zap.String("addr", s.cfg.Address))
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

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(time.Duration(s.cfg.HandshakeTimeoutS) * time.Second))
	reader := bufio.NewReaderSize(conn, s.cfg.ReadBufferBytes)

	for {
		// read length
		lenBuf := make([]byte, 4)
		if _, err := io.ReadFull(reader, lenBuf); err != nil {
			if !errors.Is(err, io.EOF) {
				s.logger.Debug("read length failed", zap.Error(err))
			}
			return
		}
		pktLen := int(binary.BigEndian.Uint32(lenBuf))
		if pktLen < protocol.HeaderLen || pktLen > 1<<20 {
			s.logger.Warn("invalid packet length", zap.Int("len", pktLen))
			return
		}

		payload := make([]byte, pktLen-4)
		if _, err := io.ReadFull(reader, payload); err != nil {
			s.logger.Debug("read payload failed", zap.Error(err))
			return
		}

		data := append(lenBuf, payload...)
		_, _, cmdID, userID, seqID, body, err := protocol.ParsePacket(data)
		if err != nil {
			s.logger.Warn("parse packet failed", zap.Error(err))
			continue
		}

		ctx := &Context{
			Server: s,
			Conn:   conn,
			CmdID:  cmdID,
			UserID: userID,
			SeqID:  seqID,
			Body:   body,
		}

		s.dispatch(ctx)
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
	if s.defaultHandle != nil {
		s.defaultHandle(ctx)
		return
	}
}

// SendResponse writes response to client.
func (s *Server) SendResponse(conn net.Conn, cmdID int32, userID uint32, body []byte) {
	resp := protocol.BuildResponse(cmdID, userID, 0, body)
	_, _ = conn.Write(resp)
}
