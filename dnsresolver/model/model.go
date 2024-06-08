package model

type Query struct {
	Header Header
	Class  Class
}
type Class struct {
	Question   string
	QueryType  uint16
	QueryClass uint16
}
type Header struct {
	Id      uint16
	Flag    uint16
	Qdcount uint16
	Ancount uint16
	Nscount uint16
	Arcount uint16
}
type Resource struct {
	Class    Class
	Ttl      uint32 //checkl
	Rdlength uint16
	Rdata    string // check
}
