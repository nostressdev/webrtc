package sdp

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	NetworkInternet = "IN"
)

const (
	TypeIPv4 = "IP4"
	TypeIPv6 = "IP6"
)

var testVectors = []*testVector{
	{
		Name: "RFC4566 Example",
		Data: `v=0
o=jdoe 2890844526 2890842807 IN IP4 10.47.16.5
s=SDP Seminar
i=A Seminar on the session description protocol
u=http://www.example.com/seminars/sdp.pdf
e=j.doe@example.com (Jane Doe)
p=+1 617 555-6011
c=IN IP4 224.2.17.12/127
b=AS:2000
t=3034423619 3042462419
r=7d 1h 0 25h
z=3034423619 -1h 3042462419 0
a=recvonly
m=audio 49170 RTP/AVP 0
m=video 51372 RTP/AVP 99 100
a=rtpmap:99 h263-1998/90000
a=rtpmap:100 H264/90000
a=rtcp-fb:100 ccm fir
a=rtcp-fb:100 nack
a=rtcp-fb:100 nack pli
a=fmtp:100 profile-level-id=42c01f;level-asymmetry-allowed=1
`,
		Session: &Session{
			Originator: &Origin{
				Username:       "jdoe",
				SessID:         2890844526,
				SessVersion:    2890842807,
				Nettype:        NetworkInternet,
				Addrtype:       TypeIPv4,
				UnicastAddress: "10.47.16.5",
			},
			SessionName:  "SDP Seminar",
			Information:  "A Seminar on the session description protocol",
			URI:          "http://www.example.com/seminars/sdp.pdf",
			Emails:       []string{"j.doe@example.com (Jane Doe)"},
			PhoneNumbers: []string{"+1 617 555-6011"},
			ConnectionData: &Connection{
				Nettype:        NetworkInternet,
				Addrtype:       TypeIPv4,
				ConnectionAddr: "224.2.17.12",
				TTL:            127,
				AddressesNum:   1,
			},
			Bandwidths: []*Bandwidth{
				{"AS", 2000},
			},
			Timings: []*Timing{
				{
					Start: 3034423619,
					Stop:  3042462419,
					RepeatTimes: []*RepeatTime{
						{
							Interval: 604800,
							Duration: 3600,
							Offsets: []int64{
								0,
								90000,
							},
						},
					},
				},
			},
			TimeZones: []*TimeZone{
				{Time: 3034423619, Offset: -3600},
				{Time: 3042462419, Offset: 0},
			},
			MediaDescs: []*MediaDesc{
				{
					Media:    "audio",
					Port:     49170,
					Proto:    []string{"RTP", "AVP"},
					PortsNum: 1,
					Fmts: []string{
						"0",
					},
				},
				{
					Media:    "video",
					Port:     51372,
					Proto:    []string{"RTP", "AVP"},
					PortsNum: 1,
					Fmts:     []string{"99", "100"},
					Attributes: []*Attribute{
						{
							Name:  "rtpmap",
							Value: "99 h263-1998/90000",
						},
						{
							Name:  "rtpmap",
							Value: "100 H264/90000",
						},
						{
							Name:  "rtcp-fb",
							Value: "100 ccm fir",
						},
						{
							Name:  "rtcp-fb",
							Value: "100 nack",
						},
						{
							Name:  "rtcp-fb",
							Value: "100 nack pli",
						},
						{
							Name:  "fmtp",
							Value: "100 profile-level-id=42c01f;level-asymmetry-allowed=1",
						},
					},
				},
			},
			Attributes: []*Attribute{
				{
					Name:  "recvonly",
					Value: " ",
				},
			},
		},
	},
	{
		Name: "Readme Example",
		Data: `v=0
o=alice 2890844526 2890844526 IN IP4 alice.example.org
s=Example
c=IN IP4 127.0.0.1
t=0 0
a=sendrecv
m=audio 10000 RTP/AVP 0 8
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
`,
		Session: &Session{
			Originator: &Origin{
				Username:       "alice",
				SessID:         2890844526,
				SessVersion:    2890844526,
				Nettype:        NetworkInternet,
				Addrtype:       TypeIPv4,
				UnicastAddress: "alice.example.org",
			},
			Attributes: []*Attribute{
				{Name: "sendrecv", Value: " "},
			},
			SessionName: "Example",
			ConnectionData: &Connection{
				Nettype:        NetworkInternet,
				Addrtype:       TypeIPv4,
				ConnectionAddr: "127.0.0.1",
				AddressesNum:   1,
			},
			Timings: []*Timing{
				{Start: 0, Stop: 0},
			},
			MediaDescs: []*MediaDesc{
				{
					Media:    "audio",
					Port:     10000,
					PortsNum: 1,
					Proto:    []string{"RTP", "AVP"},
					Fmts:     []string{"0", "8"},
					Attributes: []*Attribute{
						{
							Name:  "rtpmap",
							Value: "0 PCMU/8000",
						},
						{
							Name:  "rtpmap",
							Value: "8 PCMA/8000",
						},
					},
				},
			},
		},
	},
	{
		Name: "SCTP Example",
		Data: `v=0
o=- 0 2 IN IP4 127.0.0.1
s=-
c=IN IP4 127.0.0.1
t=0 0
m=application 10000 DTLS/SCTP 5000
a=sctpmap:5000 webrtc-datachannel 256
m=application 10000 UDP/DTLS/SCTP webrtc-datachannel
a=sctp-port:5000
`,
		Session: &Session{
			Originator: &Origin{
				Username:       "-",
				SessID:         0,
				SessVersion:    2,
				Nettype:        NetworkInternet,
				Addrtype:       TypeIPv4,
				UnicastAddress: "127.0.0.1",
			},
			SessionName: "-",
			ConnectionData: &Connection{
				Nettype:        NetworkInternet,
				Addrtype:       TypeIPv4,
				ConnectionAddr: "127.0.0.1",
				AddressesNum:   1,
			},
			Timings: []*Timing{{
				Start: 0,
				Stop:  0,
			}},
			MediaDescs: []*MediaDesc{
				{
					Media:    "application",
					Port:     10000,
					PortsNum: 1,
					Proto:    []string{"DTLS", "SCTP"},
					Fmts:     []string{"5000"},
					Attributes: []*Attribute{
						{Name: "sctpmap", Value: "5000 webrtc-datachannel 256"},
					},
				},
				{
					Media:    "application",
					Port:     10000,
					PortsNum: 1,
					Proto:    []string{"UDP", "DTLS", "SCTP"},
					Fmts:     []string{"webrtc-datachannel"},
					Attributes: []*Attribute{
						{Name: "sctp-port", Value: "5000"},
					},
				},
			},
		},
	},
}

type testVector struct {
	Name    string
	Data    string
	Session *Session
}

type T struct {
	*testing.T
}

func TestUnmarshal(t *testing.T) {
	for _, v := range testVectors {
		v := v
		t.Run(v.Name, func(inner *testing.T) {
			t := &T{inner}

			d := NewDecoder(strings.NewReader(v.Data))
			sess, err := d.Decode()
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(sess, v.Session) {
				t.Fatalf("bad Session, got: %s, expected: %s, diff: %v", dump(sess), dump(v.Session), cmp.Diff(sess, v.Session))
			}
		})
	}
}

func dump(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic(err)
	}

	return string(b)
}
