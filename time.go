package xmpp

import (
	"encoding/xml"
	"time"
)

const (
	TimeTZ  = "2006-01-02T15:04:05Z"
	TimeOld = "20060102T15:04:05"
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

// TimeRequest is XEP-0202: Time Entity query
type TimeQuery struct {
	XMLName xml.Name `xml:"urn:xmpp:time time"`
}

// String is pretty printer.
func (r *TimeReply) String() string {
	return r.Format(time.RubyDate)
}

// Format formating TimeReply with specified layout.
// see: `godoc time` for Format examples.
func (r *TimeReply) Format(layout string) string {
	if len(r.UTC) > len(TimeTZ) {
		r.UTC = r.UTC[0:len(TimeTZ)]
	}
	t, err := time.Parse(TimeTZ, r.UTC)
	if err != nil {
		return err.Error()
	}
	tl, err := time.Parse("-07:00", r.TZO)
	if err != nil {
		return err.Error()
	}
	return t.In(tl.Location()).Format(layout)
}
