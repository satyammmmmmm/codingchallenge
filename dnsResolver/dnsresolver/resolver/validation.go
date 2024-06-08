package resolver

import "github.com/satyammmmmmm/codingchallenge/tree/main/dnsResolver/dnsresolver/model"

func validatingHeaders(dnsHeader *model.Header, query *model.Query) bool {
	return true
}

func validatingClass(dnsClass *model.Class, query *model.Query) bool {
	return true
}
