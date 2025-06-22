package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

type Server struct {
	Conf     *Conf
	Handler  *Handler
	Listener net.Listener
	Quit     chan os.Signal
	Wait     *sync.WaitGroup
}

func (s *Server) Start() {
	//fmt.Println(Logo)
	var err error
	s.Listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.Conf.Ip, s.Conf.Port))
	HandleErr(err)
	Info("Listen %s", s.Listener.Addr().String())
	go s.accept()
	// 阻塞处理信号
	//signal.Notify(s.Quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	//<-s.Quit
}

func (s *Server) accept() {
	for {
		if s.Listener == nil {
			break
		}
		conn, err := s.Listener.Accept()
		HandleErr(err)
		Info("Accept %s", conn.RemoteAddr().String())
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	s.Wait.Add(1)
	defer s.Wait.Done()
	s.Handler.Handle(conn)
	err := conn.Close()
	HandleErr(err)
}

func (s *Server) Close() {
	// 关闭监听
	err := s.Listener.Close()
	HandleErr(err)
	s.Listener = nil
	// 关闭处理对象
	s.Handler.Close()
	// 已经创建的链接还是要处理完毕的
	s.Wait.Wait()
}

func NewServer(conf *Conf) *Server {
	return &Server{Conf: conf, Quit: make(chan os.Signal), Handler: NewHandler(conf), Wait: &sync.WaitGroup{}}
}
