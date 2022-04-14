// Package sdp implements Session Description Protocol (SDP), rfc4566
package sdp

type Origin struct {
	Username       string
	SessId         int64
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

type Key struct {
	Method string
	Value  string
}

type Format struct {
	Payload   uint8
	Name      string
	ClockRate int
	Channels  int
	Feedback  []string
	Params    []string
}

type MediaDisc struct {
	Media       string
	Port        int64
	PortsNum    int64
	Proto       string
	Fmt         string
	Attributes  []Attribute
	Bandwidths  []*Bandwidth
	Connections []*Connection
	Keys        []*Key
	Mode        string
	Formats     []*Format
	FormatDescr string
}

type Session struct {
	Version            int
	Originator         *Origin
	SessionName        string
	SessionInformation string
	URI                string
	Emails             []string
	PhoneNumbers       []string
	ConnectionData     *Connection
	Bandwidth          []*Bandwidth
	Timings            []*Timing
	TimeZones          []*TimeZone
	EncryptionKeys     []*EncryptionKey
	Attributes         []*Attribute
	MediaDiscs         []MediaDisc
}
