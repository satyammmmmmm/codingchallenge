package resolver

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"github.com/satyammmmmmm/codingchallenge/tree/main/dnsresolver/model"
)

func encode(h *model.Header, c *model.Class) ([]byte, error) {
	headerbytes := make([]byte, 12)
	binary.BigEndian.PutUint16(headerbytes[:2], h.Id)
	binary.BigEndian.PutUint16(headerbytes[2:4], h.Flag)
	binary.BigEndian.PutUint16(headerbytes[4:6], h.Qdcount)
	binary.BigEndian.PutUint16(headerbytes[6:8], h.Ancount)
	binary.BigEndian.PutUint16(headerbytes[8:10], h.Nscount)
	binary.BigEndian.PutUint16(headerbytes[10:12], h.Arcount)

	s := strings.Split(c.Question, ".")
	lengthofs := 0
	for i := range s {
		lengthofs += len(s[i])

	}
	size := 4 + lengthofs + len(s) + 1 // qtype+qclss+space count+null
	bodybytes := make([]byte, size)
	idx := 0
	for i := range s {
		bodybytes[idx] = uint8(len(s[i]))
		for j := 0; j < len(s[i]); j++ {
			bodybytes[idx+1+j] = s[i][j]
		}
		idx += 1 + len(s[i])

	}
	bodybytes[idx] = 0
	binary.BigEndian.PutUint16(bodybytes[size-4:size-2], c.QueryType)
	binary.BigEndian.PutUint16(bodybytes[size-2:size], c.QueryClass)
	//  bodybytes contain the domain name segments, the null termination, and the query type and class.
	encodedMessage := append(headerbytes, bodybytes...)
	return encodedMessage, nil

}
func HeaderFromDns(header []byte, offset int) (*model.Header, int, error) {
	h := model.Header{
		Id:      binary.BigEndian.Uint16(header[offset : offset+2]),
		Flag:    binary.BigEndian.Uint16(header[offset+2 : offset+4]),
		Qdcount: binary.BigEndian.Uint16(header[offset+4 : offset+6]),
		Ancount: binary.BigEndian.Uint16(header[offset+6 : offset+8]),
		Nscount: binary.BigEndian.Uint16(header[offset+8 : offset+10]),
		Arcount: binary.BigEndian.Uint16(header[offset+10 : offset+12]),
	}
	// Return header size for consistency.
	return &h, 12, nil
}

func ExtractDomainNamFromDnsResponse(buffer []byte, offset int) (string, int) {
	s := ""
	idx := offset

	for {
		length := int(buffer[idx])
		// length 192 indicates a pointer
		if length == 192 {
			//jump over the pointer
			suffix, _ := ExtractDomainNamFromDnsResponse(buffer, int(buffer[idx+1]))
			s += suffix
			idx += 2
			break
		} else {
			name := buffer[idx+1 : idx+1+length]
			idx += 1 + length
			if buffer[idx] == 0x00 {
				s += string(name)
				idx++
				break
			} else {
				s += string(name) + "."
			}
		}
	}

	return s, idx - offset
}
func decodeNSrData(buffer, rdata []byte) string {
	s := ""
	idx := 0
	for {
		length := int(rdata[idx])
		// length 192 indicates a pointer
		if length == 192 {
			// pointer to a string in the original response buffer
			suffix, _ := ExtractDomainNamFromDnsResponse(buffer, int(rdata[idx+1]))
			s += suffix
			idx += 2
			break
		} else {
			name := rdata[idx+1 : idx+1+length]
			idx += 1 + length
			if rdata[idx] == 0x00 {
				s += string(name)
				idx++
				break
			} else {
				s += string(name) + "."
			}
		}
	}
	return s
}
func decodeResource(buffer []byte, startPosition int) (*model.Resource, int, error) { //(buffer,28)

	name, size := ExtractDomainNamFromDnsResponse(buffer, startPosition)
	offset := startPosition + size

	qType := binary.BigEndian.Uint16(buffer[offset : offset+2])
	qClass := binary.BigEndian.Uint16(buffer[offset+2 : 4+offset])
	ttl := binary.BigEndian.Uint32(buffer[offset+4 : offset+8])
	rdLength := binary.BigEndian.Uint16(buffer[8+offset : 10+offset])

	rData := []byte{}
	if qType == 2 && qClass == 1 {
		rData = []byte(decodeNSrData(buffer, buffer[10+offset:10+offset+int(rdLength)]))
	} else {
		rData = buffer[10+offset : 10+uint16(offset)+rdLength]
	}
	Class :=
		model.Class{
			Question:   name,
			QueryType:  qType,
			QueryClass: qClass,
		}
	resource := model.Resource{Class, ttl, rdLength, string(rData)}

	endPosition := offset + 10 + int(rdLength)
	return &resource, endPosition - startPosition, nil
}

func ClassFromDns(buffer []byte, pointer int) (*model.Class, int, error) {
	name, size := ExtractDomainNamFromDnsResponse(buffer, pointer)
	offset := pointer + size
	class := model.Class{
		Question:   name,
		QueryType:  binary.BigEndian.Uint16(buffer[offset : offset+2]),
		QueryClass: binary.BigEndian.Uint16(buffer[offset+2 : offset+4]),
	}

	// Return size of body since it varies with domain name length.
	fmt.Println("hello=", size+4)
	return &class, size + 4, nil
}
func DomainnameResolver(domainName string, nameServer string) (string, error) {
	var header model.Header
	var query model.Query
	var class model.Class

	//define,header,class,create query for server
	header = model.Header{Id: 111, Flag: 0, Qdcount: 1, Ancount: 0, Nscount: 0, Arcount: 0}
	class = model.Class{domainName, 1, 1}
	query = model.Query{Header: header, Class: class}

	//convert into bytes(remember uint16 thing)
	message, err := encode(&query.Header, &query.Class)
	if err != nil {
		return "", err
	}
	rootServerIp := nameServer

	searchIp := make(map[string]int)
	searchIp[rootServerIp] = 1
	iplist := []string{rootServerIp}

	//logic starts here
	for len(iplist) > 0 {
		ip := iplist[0]
		iplist = iplist[1:]

		// make connection string
		connString := ip + ":53"
		//make udp connection

		connection, err := net.Dial("udp", connString)
		if err != nil {
			return "", err
		}
		defer connection.Close()

		fmt.Printf("searching server=%s for %s\n", ip, domainName)
		_, err = connection.Write(message) //send query to udp
		if err != nil {
			return "", err
		}
		dnsResponse := make([]byte, 512)      //512 is max size of dns response
		_, err = connection.Read(dnsResponse) // read dns response (see resolver.txt for sample response)
		if err != nil {
			return "", err
		}
		//RemoveHeader from dnsResponse
		//RemoveBody from dnsResponse
		//Also validate header by check ID of response and query and also body of both
		pointer := 0
		dnsHeader, size, err := HeaderFromDns(dnsResponse, pointer)
		if err != nil {
			return "", nil
		}
		valid := validatingHeaders(dnsHeader, &query)
		if valid == false {
			return "", fmt.Errorf("headers are not matching")
		}
		pointer += size
		dnsClass, size, err := ClassFromDns(dnsResponse, 12)
		if err != nil {
			return "", err
		}
		valid = validatingClass(dnsClass, &query)
		if valid == false {
			return "", fmt.Errorf("headers are not matching")
		}
		pointer += size

		for i := 0; i < int(dnsHeader.Ancount); i++ {
			ip, _, err := decodeResource(dnsResponse, pointer)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%d.%d.%d.%d", ip.Rdata[0], ip.Rdata[1], ip.Rdata[2], ip.Rdata[3]), nil

		}

		authoritySection := make([]*model.Resource, 0)
		for i := 0; i < int(dnsHeader.Nscount); i++ {
			record, size, err := decodeResource(dnsResponse, pointer)
			if err != nil {
				return "", err
			}
			authoritySection = append(authoritySection, record)
			pointer += size
		}
		additionalRecords := make([]*model.Resource, 0)
		for i := 0; i < int(dnsHeader.Arcount); i++ {
			additional, size, err := decodeResource(dnsResponse, pointer)
			if err != nil {
				return "", err
			}
			additionalRecords = append(additionalRecords, additional)
			pointer += size
		}

		for i := range additionalRecords {
			// We have ipv4 address for server that can help resolve the query.
			ar := additionalRecords[i]
			if ar.Class.QueryType == 1 && ar.Class.QueryClass == 1 && ar.Rdlength == 4 {
				newIP := fmt.Sprintf("%d.%d.%d.%d", ar.Rdata[0], ar.Rdata[1], ar.Rdata[2], ar.Rdata[3])
				if _, exists := searchIp[newIP]; !exists {
					iplist = append(iplist, newIP)
					searchIp[newIP] = 1
				}
			}
		}

		// Need to resolve name server's ip address to continue.
		if len(iplist) == 0 && len(authoritySection) > 0 {
			fmt.Println("Querying for name server ip.")
			nameServer, err := DomainnameResolver(string(authoritySection[0].Rdata), rootServerIp)
			if err != nil {
				return "", err
			}
			iplist = append(iplist, nameServer)
		}
	}
	return "", fmt.Errorf("Failed to resolve this domain name.")

}
