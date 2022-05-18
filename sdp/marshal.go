package sdp

import (
	"io"
	"strconv"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(s *Session) {
	e.encodeSession(s)
}

func (e *Encoder) writeInt64(v int64) *Encoder {
	e.w.Write([]byte(strconv.FormatInt(v, 10)))
	return e
}

func (e *Encoder) writeInt(v int) *Encoder {
	return e.writeInt64(int64(v))
}

func (e *Encoder) writeString(v string) *Encoder {
	e.w.Write([]byte(v))
	return e
}

func (e *Encoder) writeChar(char byte) *Encoder {
	e.w.Write([]byte{char})
	return e
}

func (e *Encoder) writeNewline() *Encoder {
	e.w.Write([]byte{'\n'})
	return e
}

func (e *Encoder) writeSpace() *Encoder {
	e.w.Write([]byte{' '})
	return e
}

func (e *Encoder) writeField(field byte) *Encoder {
	return e.writeChar(field).writeChar('=')
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
	e.writeField(VersionField).writeInt(version).writeNewline()
}

func (e *Encoder) encodeOriginator(originator *Origin) {
	e.writeField(OriginField).writeString(originator.Username).writeSpace()
	e.writeInt64(originator.SessID).writeSpace().writeInt64(originator.SessVersion).writeSpace()
	e.writeString(originator.Nettype).writeSpace().writeString(originator.Addrtype).writeSpace()
	e.writeString(originator.UnicastAddress).writeNewline()
}

func (e *Encoder) encodeSessionName(name string) {
	e.writeField(SessionNameField).writeString(name).writeNewline()
}

func (e *Encoder) encodeURI(URI string) {
	e.writeField(URIField).writeString(URI).writeNewline()
}

func (e *Encoder) encodeInformation(info string) {
	e.writeField(SessionInfoField).writeString(info).writeNewline()
}

func (e *Encoder) encodeEmail(email string) {
	e.writeField(EmailField).writeString(email).writeNewline()
}

func (e *Encoder) encodeEmails(emails []string) {
	for _, email := range emails {
		e.encodeEmail(email)
	}
}

func (e *Encoder) encodePhoneNumber(phone string) {
	e.writeField(PhoneNumberField).writeString(phone).writeNewline()
}

func (e *Encoder) encodePhoneNumbers(phones []string) {
	for _, phone := range phones {
		e.encodePhoneNumber(phone)
	}
}

func (e *Encoder) encodeConnection(connection *Connection) {
	e.writeField(ConnectionDataField).writeString(connection.Nettype).writeSpace().writeString(connection.Addrtype).writeSpace()
	e.writeString(connection.ConnectionAddr)
	if connection.TTL > 0 {
		e.writeChar('/').writeInt64(connection.TTL)
	}
	e.writeChar('/').writeInt64(connection.AddressesNum).writeNewline()
}

func (e *Encoder) encodeConnections(connections []*Connection) {
	for _, connection := range connections {
		e.encodeConnection(connection)
	}
}

func (e *Encoder) encodeBandwidth(bandwidth *Bandwidth) {
	e.writeField(BandwidthField).writeString(bandwidth.Type).writeChar(':').writeInt(bandwidth.Value).writeNewline()
}

func (e *Encoder) encodeBandwidths(bandwidths []*Bandwidth) {
	for _, bandwidth := range bandwidths {
		e.encodeBandwidth(bandwidth)
	}
}

func (e *Encoder) encodeRepeatTime(time *RepeatTime) {
	e.writeField(RepeatTimeField).writeInt64(time.Interval).writeSpace().writeInt64(time.Duration)

	if len(time.Offsets) > 0 {
		e.writeSpace()
	}
	for i, offset := range time.Offsets {
		e.writeInt64(offset)

		if i+1 != len(time.Offsets) {
			e.writeSpace()
		}
	}
}

func (e *Encoder) encodeRepeatTimes(times []*RepeatTime) {
	for _, time := range times {
		e.encodeRepeatTime(time)
	}
}

func (e *Encoder) encodeTiming(timing *Timing) {
	e.writeField(TimingField).writeInt64(timing.Start).writeSpace().writeInt64(timing.Stop)

	if timing.RepeatTimes != nil {
		e.writeNewline()
		e.encodeRepeatTimes(timing.RepeatTimes)
	}
	e.writeNewline()
}

func (e *Encoder) encodeTimings(timings []*Timing) {
	for _, timing := range timings {
		e.encodeTiming(timing)
	}
}

func (e *Encoder) encodeTimeZones(zones []*TimeZone) {
	e.writeField(TimeZoneField)

	for i, zone := range zones {
		e.writeInt64(zone.Time).writeSpace().writeInt64(zone.Offset)
		if i+1 != len(zones) {
			e.writeSpace()
		}
	}

	e.writeNewline()
}

func (e *Encoder) encodeEncryptionKey(key *EncryptionKey) {
	e.writeField(EncryptionKeyField).writeString(key.Method)

	if key.Value != " " {
		e.writeChar(':').writeString(key.Value)
	}
	e.writeNewline()
}

func (e *Encoder) encodeEncryptionKeys(encryptionKeys []*EncryptionKey) {
	for _, key := range encryptionKeys {
		e.encodeEncryptionKey(key)
	}
}

func (e *Encoder) encodeAttribute(attribute *Attribute) {
	e.writeField(AttributeField).writeString(attribute.Name)
	if attribute.Value != " " && attribute.Value != "" {
		e.writeChar(':').writeString(attribute.Value)
	}
	e.writeNewline()
}

func (e *Encoder) encodeAttributes(attributes []*Attribute) {
	for _, attribute := range attributes {
		e.encodeAttribute(attribute)
	}
}

func (e *Encoder) encodeMediaDesc(desc *MediaDesc) {
	e.writeField(MediaDescField).writeString(desc.Media).writeSpace().writeInt64(desc.Port).writeChar('/').writeInt64(desc.PortsNum).writeSpace()
	for i, proto := range desc.Proto {
		e.writeString(proto)
		if i+1 != len(desc.Proto) {
			e.writeChar('/')
		}
	}
	if len(desc.Fmts) != 0 {
		e.writeSpace()
		for i, fmt := range desc.Fmts {
			e.writeString(fmt)
			if i+1 != len(desc.Fmts) {
				e.writeSpace()
			}
		}
	}

	e.writeNewline()

	if desc.Information != "" {
		e.writeField(SessionInfoField).writeString(desc.Information).writeNewline()
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
