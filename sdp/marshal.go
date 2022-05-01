package sdp

import (
	"io"
	"strconv"
)

type Encoder struct {
	data []byte
	w    io.Writer
}

const (
	defaultBufferSize = 1024
)

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w, data: make([]byte, 0, defaultBufferSize)}
}

func (e *Encoder) Encode(s *Session) error {
	e.data = e.data[:0]
	e.data = append(e.data, e.encodeSession(s)...)
	return e.Flush()
}

func (e *Encoder) Flush() error {
	if data := e.data; len(data) > 0 {
		written, err := e.w.Write(data)
		e.data = data[:copy(data, data[written:])]
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeSession(s *Session) string {
	var res string

	res += e.encodeVersion(s.Version)
	res += e.encodeOriginator(s.Originator)
	res += e.encodeSessionName(s.SessionName)

	if s.Information != "" {
		res += e.encodeInformation(s.Information)
	}
	if s.URI != "" {
		res += e.encodeURI(s.URI)
	}
	if s.Emails != nil {
		res += e.encodeEmails(s.Emails)
	}
	if s.PhoneNumbers != nil {
		res += e.encodePhoneNumbers(s.PhoneNumbers)
	}
	if s.ConnectionData != nil {
		res += e.encodeConnection(s.ConnectionData)
	}
	if s.Bandwidths != nil {
		res += e.encodeBandwidths(s.Bandwidths)
	}
	if s.Timings != nil {
		res += e.encodeTimings(s.Timings)
	}
	if s.TimeZones != nil {
		res += e.encodeTimeZones(s.TimeZones)
	}
	if s.EncryptionKeys != nil {
		res += e.encodeEncryptionKeys(s.EncryptionKeys)
	}
	if s.Attributes != nil {
		res += e.encodeAttributes(s.Attributes)
	}
	if s.MediaDescs != nil {
		res += e.encodeMediaDescs(s.MediaDescs)
	}

	return res
}

func (e *Encoder) encodeVersion(version int) string {
	return "v=" + strconv.Itoa(version) + "\n"
}

func (e *Encoder) encodeOriginator(originator *Origin) string {
	return "o=" + originator.Username + " " + strconv.FormatInt(originator.SessID, 10) + " " + strconv.FormatInt(originator.SessVersion, 10) + " " + originator.Nettype + " " + originator.Addrtype + " " + originator.UnicastAddress + "\n"
}

func (e *Encoder) encodeSessionName(name string) string {
	return "s=" + name + "\n"
}

func (e *Encoder) encodeURI(URI string) string {
	return "u=" + URI + "\n"
}

func (e *Encoder) encodeInformation(info string) string {
	return "i=" + info + "\n"
}

func (e *Encoder) encodeEmail(email string) string {
	return "e=" + email + "\n"
}

func (e *Encoder) encodeEmails(emails []string) string {
	var res string
	for _, email := range emails {
		res += e.encodeEmail(email)
	}
	return res
}

func (e *Encoder) encodePhoneNumber(phone string) string {
	return "p=" + phone + "\n"
}

func (e *Encoder) encodePhoneNumbers(phones []string) string {
	var res string
	for _, phone := range phones {
		res += e.encodePhoneNumber(phone)
	}
	return res
}

func (e *Encoder) encodeConnection(connection *Connection) string {
	var res string

	res += "c=" + connection.Nettype + " " + connection.Addrtype + " " + connection.ConnectionAddr
	if connection.TTL > 0 {
		res += "/" + strconv.FormatInt(connection.TTL, 10)
	}
	res += "/" + strconv.FormatInt(connection.AddressesNum, 10) + "\n"
	return res
}

func (e *Encoder) encodeConnections(connections []*Connection) string {
	var res string
	for _, connection := range connections {
		res += e.encodeConnection(connection)
	}
	return res
}

func (e *Encoder) encodeBandwidth(bandwidth *Bandwidth) string {
	return "b=" + bandwidth.Type + ":" + strconv.Itoa(bandwidth.Value) + "\n"
}

func (e *Encoder) encodeBandwidths(bandwidths []*Bandwidth) string {
	var res string
	for _, bandwidth := range bandwidths {
		res += e.encodeBandwidth(bandwidth)
	}
	return res
}

func (e *Encoder) encodeRepeatTime(time *RepeatTime) string {
	var res string

	res += "r=" + strconv.FormatInt(time.Interval, 10) + " " + strconv.FormatInt(time.Duration, 10)

	if len(time.Offsets) > 0 {
		res += " "
	}
	for i, offset := range time.Offsets {
		res += strconv.FormatInt(offset, 10)

		if i+1 != len(time.Offsets) {
			res += " "
		}
	}

	return res
}

func (e *Encoder) encodeRepeatTimes(times []*RepeatTime) string {
	var res string
	for _, time := range times {
		res += e.encodeRepeatTime(time)
	}
	return res
}

func (e *Encoder) encodeTiming(timing *Timing) string {
	var res string
	res += "t=" + strconv.FormatInt(timing.Start, 10) + " " + strconv.FormatInt(timing.Stop, 10)

	if timing.RepeatTimes != nil {
		res += "\n"
		res += e.encodeRepeatTimes(timing.RepeatTimes)
	}
	res += "\n"

	return res
}

func (e *Encoder) encodeTimings(timings []*Timing) string {
	var res string
	for _, timing := range timings {
		res += e.encodeTiming(timing)
	}
	return res
}

func (e *Encoder) encodeTimeZones(zones []*TimeZone) string {
	var res string

	res += "z="

	for i, zone := range zones {
		res += strconv.FormatInt(zone.Time, 10) + " " + strconv.FormatInt(zone.Offset, 10)
		if i+1 != len(zones) {
			res += " "
		}
	}

	res += "\n"

	return res
}

func (e *Encoder) encodeEncryptionKey(key *EncryptionKey) string {
	var res string
	res += "k=" + key.Method
	if key.Value != " " {
		res += ":" + key.Value
	}
	res += "\n"
	return res
}

func (e *Encoder) encodeEncryptionKeys(encryptionKeys []*EncryptionKey) string {
	var res string
	for _, key := range encryptionKeys {
		res += e.encodeEncryptionKey(key)
	}
	return res
}

func (e *Encoder) encodeAttribute(attribute *Attribute) string {
	var res string
	res += "a=" + attribute.Name
	if attribute.Value != " " {
		res += ":" + attribute.Value
	}
	res += "\n"
	return res
}

func (e *Encoder) encodeAttributes(attributes []*Attribute) string {
	var res string
	for _, attribute := range attributes {
		res += e.encodeAttribute(attribute)
	}
	return res
}

func (e *Encoder) encodeMediaDesc(desc *MediaDesc) string {
	var res string
	res += "m=" + desc.Media + " " + strconv.FormatInt(desc.Port, 10) + "/" + strconv.FormatInt(desc.PortsNum, 10) + " "
	for i, proto := range desc.Proto {
		res += proto
		if i+1 != len(desc.Proto) {
			res += "/"
		}
	}
	res += " "
	for i, fmt := range desc.Fmts {
		res += fmt
		if i+1 != len(desc.Fmts) {
			res += " "
		}
	}

	res += "\n"

	if desc.Information != "" {
		res += desc.Information + "\n"
	}
	if desc.Connections != nil {
		res += e.encodeConnections(desc.Connections)
	}
	if desc.Bandwidths != nil {
		res += e.encodeBandwidths(desc.Bandwidths)
	}
	if desc.EncryptionKeys != nil {
		res += e.encodeEncryptionKeys(desc.EncryptionKeys)
	}
	if desc.Attributes != nil {
		res += e.encodeAttributes(desc.Attributes)
	}

	return res
}

func (e *Encoder) encodeMediaDescs(descs []*MediaDesc) string {
	var res string
	for _, desc := range descs {
		res += e.encodeMediaDesc(desc)
	}
	return res
}
