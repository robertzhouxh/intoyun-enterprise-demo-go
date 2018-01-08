package main

type Server struct {
	operator Operator
}

func NewServer(o Operator) *Server {
	s := new(Server)
	s.operator = o
	return s
}
