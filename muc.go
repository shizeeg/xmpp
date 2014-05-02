package xmpp

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
)

const (
	NsMUC       = "http://jabber.org/protocol/muc"
	NsMUCUser   = "http://jabber.org/protocol/muc#user"
	NsDiscoInfo = "http://jabber.org/protocol/disco#info"
)

const (
	SELF   = "110"
	BAN    = "301"
	RENAME = "303"
	KICK   = "307"
)

type Status struct {
	// XMLName xml.Name `xml:"status"`
	Code string `xml:"code,attr"`
}
type Actor struct {
	Nick string `xml:"nick,attr,omitempty"`
	JID  string `xml:"jid,attr,omitempty"`
}

type Item struct {
	// XMLName xml.Name `xml:"item"`
	// owner, admin, member, outcast, none
	Affil string `xml:"affiliation,attr,omitempty"`
	// moderator, participant, visitor, none
	Role   string `xml:"role,attr,omitempty"`
	JID    string `xml:"jid,attr,omitempty"`
	Nick   string `xml:"nick,attr,omitempty"`
	Reason string `xml:"reason,omitempty"`
	Actor  Actor  `xml:"actor,omitempty"`
}

type X struct {
	XMLName  xml.Name `xml:"http://jabber.org/protocol/muc#user x"`
	Items    []Item   `xml:"item,omitempty"`
	Statuses []Status `xml:"status,omitempty"`
	Decline  Reason   `xml:"decline,omitempty"`
	Invite   Reason   `xml:"invite,omitempty"`
	Destroy  XDestroy `xml:"destroy,omitempty"`
	Password string   `xml:"password,omitempty"`
}

type Photo struct {
	XMLName xml.Name `xml:"vcard-temp:x:update x"`
	Photo   string   `xml:"photo,omitempty"`
}

// Reason common stanza for invite/decline
type Reason struct {
	From   string `xml:"from,attr,omitempty"`
	To     string `xml:"to,attr,omitempty"`
	Reason string `xml:"reason,omitempty"`
}

type XDestroy struct {
	JID    string `xml:"jid,attr,omitempty"`
	Reason string `xml:"reason,omitempty"`
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

// IsCode checks if MUCPresence contains given code
func (p *MUCPresence) IsCode(code string) bool {
	if len(p.X) == 0 {
		return false
	}
	for _, x := range p.X {
		for _, xs := range x.Statuses {
			if xs.Code == code {
				return true
			}
		}
	}
	return false
}

func (p *MUCPresence) GetAffilRole() (affil, role string, err error) {
	if len(p.X) == 0 {
		return "", "", errors.New("no <x /> subitem!")
	}
	for _, x := range p.X {
		for _, xi := range x.Items {
			affil = xi.Affil
			role = xi.Role
			return
		}
	}
	return "", "", errors.New("no affil/role info")
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
	_, err := fmt.Fprint(c.out, stanza)
	return err
}

// LeaveMUC leaves the conference.
func (c *Conn) LeaveMUC(confFullJID, status string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, err := fmt.Fprintf(c.out,
		"<presence from='%s' to='%s' type='unavailable'",
		xmlEscape(c.jid), xmlEscape(confFullJID))

	if err != nil {
		return err
	}
	if len(status) > 0 {
		_, err = fmt.Fprint(c.out, ">\n<status>"+xmlEscape(status)+
			"</status>\n</presence>")
		return err
	}
	_, err = fmt.Fprint(c.out, " />")
	return err
}

// DirectInviteMUC sent invite http://xmpp.org/extensions/xep-0249.html
func (c *Conn) DirectInviteMUC(to, jid, password, reason string) error {
	if len(password) > 0 {
		password = "password='" + password + "'"
	}

	if len(reason) > 0 {
		reason = "reason='" + xmlEscape(reason) + "'"
	}

	invite := fmt.Sprintf("<message from='%s' to='%s'>"+
		"<x xmlns='jabber:x:conference'"+
		"\n jid='%s'"+
		"\n "+password+
		"\n "+reason+" /></message>",
		xmlEscape(c.jid), xmlEscape(to), xmlEscape(jid))
	_, err := fmt.Fprint(c.out, invite)
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
	_, err := fmt.Fprint(c.out, stanza)
	return err
}
