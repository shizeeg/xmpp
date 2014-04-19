package xmpp

import (
	"encoding/xml"
)

// TimeReply is XEP-0202 reply stanza
type TimeReply struct {
	XMLName xml.Name `xml:"urn:xmpp:time time"`
	TZO     string   `xml:"tzo"`
	UTC     string   `xml:"utc"`
}

// TimeReplyOld is XEP-0090 stanza
type TimeReplyOld struct {
	XMLName xml.Name `xml:"jabber:iq:time query"`
	UTC     string   `xml:"utc"`
	TZ      string   `xml:"tz"`
	Display string   `xml:"display"`
}
