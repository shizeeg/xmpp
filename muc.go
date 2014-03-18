package xmpp

import (
	"encoding/xml"
	"fmt"
	"strings"
)

const (
	NsMUC     = "http://jabber.org/protocol/muc"
	NsMUCUser = "http://jabber.org/protocol/muc#user"
)

type Status struct {
	XMLName xml.Name `xml:"status"`
	Code    string   `xml:"code,attr"`
}

type Item struct {
	XMLName xml.Name `xml:"item"`
	// owner, admin, member, outcast, none
	Affiliation string `xml:"affiliation,attr,omitempty"`
	// moderator, participant, visitor, none
	Role string `xml:"role,attr,omitempty"`
	JID  string `xml:"jid,attr,omitempty"`
}

type X struct {
	XMLName  xml.Name `xml:"http://jabber.org/protocol/muc#user x"`
	Items    []Item   `xml:"item,omitempty"`
	Statuses []Status `xml:"status,omitempty"`
}

type Photo struct {
	XMLName xml.Name `xml:"vcard-temp:x:update x"`
	Photo   string   `xml:"photo,omitempty"`
}

// http://xmpp.org/extensions/xep-0045.html
// <presence />
type MUCPresence struct {
	XMLName xml.Name `xml:"presence"`
	Lang    string   `xml:"lang,attr,omitempty"`
	From    string   `xml:"from,attr,omitempty"`
	To      string   `xml:"to,attr,omitempty"`
	Id      string   `xml:"id.attr,omitempty"`
	Type    string   `xml:"type,attr,omitempty"`

	X     []X   `xml:"http://jabber.org/protocol/muc#user x,omitempty"`
	Photo Photo `xml:"vcard-temp:x:update x"` // http://xmpp.org/extensions/xep-0153.html

	Show     string       `xml:"show,omitempty"`   // away, chat, dnd, xa
	Status   string       `xml:"status,omitempty"` // sb []clientText
	Priority string       `xml:"priority,omitempty"`
	Caps     *ClientCaps  `xml:"c"`
	Error    *ClientError `xml:"error"`
}

// JoinMUC joins to a given conference with nick and optional password
// http://xmpp.org/extensions/xep-0045.html#bizrules-presence
func (c *Conn) JoinMUC(to, nick, password string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	// remove resource from jid
	parts := strings.SplitN(to, "/", 2)
	to = parts[0]
	if len(nick) == 0 {
		if len(parts) == 2 {
			nick = parts[1]
		} else { // if nick empty & bare jid, set nick to login
			nick = strings.SplitN(c.jid, "@", 2)[0]
		}
	}
	var pass string
	if len(password) > 0 {
		pass = "<password>" + password + "</password>"
	}
	stanza := fmt.Sprintf("<presence from='%s' to='%s/%s'>"+
		"\n  <x xmlns='%s'>"+
		"\n    <history maxchars='0' />"+
		"\n    "+pass+
		"\n  </x>"+
		"\n</presence>",
		xmlEscape(c.jid), xmlEscape(to), xmlEscape(nick), NsMUC)
	_, err := fmt.Fprintf(c.out, stanza)
	return err
}

// SendMUC sends a message to the given conference with specified type (chat or groupchat).
func (c *Conn) SendMUC(to, typ, msg string) error {
	if typ == "" {
		typ = "groupchat"
	}
	cookie := c.getCookie()
	stanza := fmt.Sprintf("<message from='%s' to='%s' type='%s' id='%x'><body>%s</body></message>",
		xmlEscape(c.jid), xmlEscape(to), xmlEscape(typ), cookie, xmlEscape(msg))
	_, err := fmt.Fprintf(c.out, stanza)
	return err
}
