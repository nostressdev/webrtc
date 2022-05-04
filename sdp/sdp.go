// Package sdp implements Session Description Protocol (SDP), rfc4566
package sdp

type Origin struct {
	Username       string
	SessID         int64
	SessVersion    int64
	Nettype        string
	Addrtype       string
	UnicastAddress string
}

type Connection struct {
	Nettype        string
	Addrtype       string
	ConnectionAddr string
	TTL            int64
	AddressesNum   int64
}

type Bandwidth struct {
	Type  string
	Value int
}

type Timing struct {
	Start       int64
	Stop        int64
	RepeatTimes []*RepeatTime
}

type RepeatTime struct {
	Interval int64
	Duration int64
	Offsets  []int64
}

type TimeZone struct {
	Time   int64
	Offset int64
}

type EncryptionKey struct {
	Method string
	Value  string
}

type Attribute struct {
	Name  string
	Value string
}

type MediaDesc struct {
	Media          string
	Information    string
	Port           int64
	PortsNum       int64
	Proto          []string
	Fmts           []string
	Attributes     []*Attribute
	Bandwidths     []*Bandwidth
	Connections    []*Connection
	EncryptionKeys []*EncryptionKey
}

type Session struct {
	Version        int
	Information    string
	Originator     *Origin
	SessionName    string
	URI            string
	Emails         []string
	PhoneNumbers   []string
	ConnectionData *Connection
	Bandwidths     []*Bandwidth
	Timings        []*Timing
	TimeZones      []*TimeZone
	EncryptionKeys []*EncryptionKey
	Attributes     []*Attribute
	MediaDescs     []*MediaDesc
}

const (
	NetworkInternet = "IN"
)

const (
	TypeIPv4 = "IP4"
	TypeIPv6 = "IP6"
)

const (
	VersionField        = 'v'
	OriginField         = 'o'
	SessionNameField    = 's'
	SessionInfoField    = 'i'
	URIField            = 'u'
	EmailField          = 'e'
	PhoneNumberField    = 'p'
	ConnectionDataField = 'c'
	BandwidthField      = 'b'
	TimingField         = 't'
	RepeatTimeField     = 'r'
	TimeZoneField       = 'z'
	EncryptionKeyField  = 'k'
	AttributeField      = 'a'
	MediaDescField      = 'm'
)

const (
	DayShorthand    = 'd'
	HourShorthand   = 'h'
	MinuteShorthand = 'm'
	SecondShorthand = 's'
)
