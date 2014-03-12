package xmpp

import (
	"encoding/xml"
	"fmt"
)

type PingQuery struct {
	XMLName xml.Name	`xml:"urn:xmpp:ping ping"`
}

func (c *Conn) KeepAlive() {
	fmt.Fprint(c.out, " ")
}
