package sdp

import (
	"io"
	"strconv"
	"sync"
)

type buffer struct {
	data []byte
}

func (b *buffer) writeInt64(v int64) *buffer {
	b.data = strconv.AppendInt(b.data, v, 10)
	return b
}

func (b *buffer) writeInt(v int) *buffer {
	return b.writeInt64(int64(v))
}

func (b *buffer) writeString(v string) *buffer {
	b.data = append(b.data, v...)
	return b
}

func (b *buffer) writeChar(char byte) *buffer {
	b.data = append(b.data, char)
	return b
}

func (b *buffer) writeNewline() *buffer {
	b.data = append(b.data, '\n')
	return b
}

func (b *buffer) writeSpace() *buffer {
	b.data = append(b.data, ' ')
	return b
}

var bufferPool = sync.Pool{
	New: func() interface{} { return &buffer{} },
}

type Encoder struct {
	buffer *buffer
	w      io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w, buffer: bufferPool.Get().(*buffer)}
}

func (e *Encoder) Encode(s *Session) error {
	e.encodeSession(s)
	return e.Flush()
}

func (e *Encoder) Flush() error {
	written := 0

	for written < len(e.buffer.data) {
		w, err := e.w.Write(e.buffer.data[written:])
		if err != nil {
			return err
		}
		written += w
	}

	e.buffer.data = e.buffer.data[:0]
	bufferPool.Put(e.buffer)

	return nil
}

func (e *Encoder) encodeSession(s *Session) {
	e.encodeVersion(s.Version)
	e.encodeOriginator(s.Originator)
	e.encodeSessionName(s.SessionName)

	if s.Information != "" {
		e.encodeInformation(s.Information)
	}
	if s.URI != "" {
		e.encodeURI(s.URI)
	}
	if s.Emails != nil {
		e.encodeEmails(s.Emails)
	}
	if s.PhoneNumbers != nil {
		e.encodePhoneNumbers(s.PhoneNumbers)
	}
	if s.ConnectionData != nil {
		e.encodeConnection(s.ConnectionData)
	}
	if s.Bandwidths != nil {
		e.encodeBandwidths(s.Bandwidths)
	}
	if s.Timings != nil {
		e.encodeTimings(s.Timings)
	}
	if s.TimeZones != nil {
		e.encodeTimeZones(s.TimeZones)
	}
	if s.EncryptionKeys != nil {
		e.encodeEncryptionKeys(s.EncryptionKeys)
	}
	if s.Attributes != nil {
		e.encodeAttributes(s.Attributes)
	}
	if s.MediaDescs != nil {
		e.encodeMediaDescs(s.MediaDescs)
	}
}

func (e *Encoder) encodeVersion(version int) {
	e.buffer.writeString("v=").writeInt(version).writeNewline()
}

func (e *Encoder) encodeOriginator(originator *Origin) {
	e.buffer.writeString("o=").writeString(originator.Username).writeSpace().writeInt64(originator.SessID).writeSpace().writeInt64(originator.SessVersion).writeSpace().writeString(originator.Nettype).writeSpace().writeString(originator.Addrtype).writeSpace().writeString(originator.UnicastAddress).writeNewline()
}

func (e *Encoder) encodeSessionName(name string) {
	e.buffer.writeString("s=").writeString(name).writeNewline()
}

func (e *Encoder) encodeURI(URI string) {
	e.buffer.writeString("u=").writeString(URI).writeNewline()
}

func (e *Encoder) encodeInformation(info string) {
	e.buffer.writeString("i=").writeString(info).writeNewline()
}

func (e *Encoder) encodeEmail(email string) {
	e.buffer.writeString("e=").writeString(email).writeNewline()
}

func (e *Encoder) encodeEmails(emails []string) {
	for _, email := range emails {
		e.encodeEmail(email)
	}
}

func (e *Encoder) encodePhoneNumber(phone string) {
	e.buffer.writeString("p=").writeString(phone).writeNewline()
}

func (e *Encoder) encodePhoneNumbers(phones []string) {
	for _, phone := range phones {
		e.encodePhoneNumber(phone)
	}
}

func (e *Encoder) encodeConnection(connection *Connection) {
	e.buffer.writeString("c=").writeString(connection.Nettype).writeSpace().writeString(connection.Addrtype).writeSpace().writeString(connection.ConnectionAddr)
	if connection.TTL > 0 {
		e.buffer.writeChar('/').writeInt64(connection.TTL)
	}
	e.buffer.writeChar('/').writeInt64(connection.AddressesNum).writeNewline()
}

func (e *Encoder) encodeConnections(connections []*Connection) {
	for _, connection := range connections {
		e.encodeConnection(connection)
	}
}

func (e *Encoder) encodeBandwidth(bandwidth *Bandwidth) {
	e.buffer.writeString("b=").writeString(bandwidth.Type).writeChar(':').writeInt(bandwidth.Value).writeNewline()
}

func (e *Encoder) encodeBandwidths(bandwidths []*Bandwidth) {
	for _, bandwidth := range bandwidths {
		e.encodeBandwidth(bandwidth)
	}
}

func (e *Encoder) encodeRepeatTime(time *RepeatTime) {
	e.buffer.writeString("r=").writeInt64(time.Interval).writeSpace().writeInt64(time.Duration)

	if len(time.Offsets) > 0 {
		e.buffer.writeSpace()
	}
	for i, offset := range time.Offsets {
		e.buffer.writeInt64(offset)

		if i+1 != len(time.Offsets) {
			e.buffer.writeSpace()
		}
	}
}

func (e *Encoder) encodeRepeatTimes(times []*RepeatTime) {
	for _, time := range times {
		e.encodeRepeatTime(time)
	}
}

func (e *Encoder) encodeTiming(timing *Timing) {
	e.buffer.writeString("t=").writeInt64(timing.Start).writeSpace().writeInt64(timing.Stop)

	if timing.RepeatTimes != nil {
		e.buffer.writeNewline()
		e.encodeRepeatTimes(timing.RepeatTimes)
	}
	e.buffer.writeNewline()
}

func (e *Encoder) encodeTimings(timings []*Timing) {
	for _, timing := range timings {
		e.encodeTiming(timing)
	}
}

func (e *Encoder) encodeTimeZones(zones []*TimeZone) {
	e.buffer.writeString("z=")

	for i, zone := range zones {
		e.buffer.writeInt64(zone.Time).writeSpace().writeInt64(zone.Offset)
		if i+1 != len(zones) {
			e.buffer.writeSpace()
		}
	}

	e.buffer.writeNewline()
}

func (e *Encoder) encodeEncryptionKey(key *EncryptionKey) {
	e.buffer.writeString("k=").writeString(key.Method)

	if key.Value != " " {
		e.buffer.writeChar(':').writeString(key.Value)
	}
	e.buffer.writeNewline()
}

func (e *Encoder) encodeEncryptionKeys(encryptionKeys []*EncryptionKey) {
	for _, key := range encryptionKeys {
		e.encodeEncryptionKey(key)
	}
}

func (e *Encoder) encodeAttribute(attribute *Attribute) {
	e.buffer.writeString("a=").writeString(attribute.Name)
	if attribute.Value != " " {
		e.buffer.writeChar(':').writeString(attribute.Value)
	}
	e.buffer.writeNewline()
}

func (e *Encoder) encodeAttributes(attributes []*Attribute) {
	for _, attribute := range attributes {
		e.encodeAttribute(attribute)
	}
}

func (e *Encoder) encodeMediaDesc(desc *MediaDesc) {
	e.buffer.writeString("m=").writeString(desc.Media).writeSpace().writeInt64(desc.Port).writeChar('/').writeInt64(desc.PortsNum).writeSpace()
	for i, proto := range desc.Proto {
		e.buffer.writeString(proto)
		if i+1 != len(desc.Proto) {
			e.buffer.writeChar('/')
		}
	}
	e.buffer.writeSpace()
	for i, fmt := range desc.Fmts {
		e.buffer.writeString(fmt)
		if i+1 != len(desc.Fmts) {
			e.buffer.writeSpace()
		}
	}

	e.buffer.writeNewline()

	if desc.Information != "" {
		e.buffer.writeString("i=").writeString(desc.Information).writeNewline()
	}
	if desc.Connections != nil {
		e.encodeConnections(desc.Connections)
	}
	if desc.Bandwidths != nil {
		e.encodeBandwidths(desc.Bandwidths)
	}
	if desc.EncryptionKeys != nil {
		e.encodeEncryptionKeys(desc.EncryptionKeys)
	}
	if desc.Attributes != nil {
		e.encodeAttributes(desc.Attributes)
	}
}

func (e *Encoder) encodeMediaDescs(descs []*MediaDesc) {
	for _, desc := range descs {
		e.encodeMediaDesc(desc)
	}
}
