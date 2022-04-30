package sdp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

var sdpTime = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

type Decoder struct {
	r io.Reader
	s *Session
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r, s: &Session{}}
}

func (d *Decoder) parseVersion(value string) (int, error) {
	version, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("Wrong version number: %v", value)
	}
	return version, nil
}

func (d *Decoder) parseTime(value string) (num int64, err error) {
	multiplyer := d.timeShorthandToSeconds(value[len(value)-1])
	if multiplyer > 0 {
		num, err = strconv.ParseInt(value[:len(value)-1], 10, 64)
	} else {
		multiplyer = 1
		num, err = strconv.ParseInt(value, 10, 64)
	}
	if err != nil {
		return 0, fmt.Errorf("Error while parsing time: %v", err)
	}
	return num * multiplyer, nil
}

func (d *Decoder) timeShorthandToSeconds(b byte) int64 {
	switch b {
	case 'd':
		return 86400
	case 'h':
		return 3600
	case 'm':
		return 60
	case 's':
		return 1
	default:
		return 0
	}
}

func (d *Decoder) parseEncryptionKey(value string) (*EncryptionKey, error) {
	var key EncryptionKey

	fields := strings.Split(value, ":")
	key.Method = fields[0]

	if len(fields) == 1 {
		key.Value = " "
	} else if len(fields) == 2 {
		key.Value = fields[1]
	} else {
		return nil, fmt.Errorf("Wrong encryption key format")
	}

	return &key, nil
}

func (d *Decoder) parseAttribute(value string) (*Attribute, error) {
	var att Attribute

	fields := strings.Split(value, ":")
	att.Name = fields[0]

	if len(fields) == 1 {
		att.Value = " "
	} else if len(fields) == 2 {
		att.Value = fields[1]
	} else {
		return nil, fmt.Errorf("Wrong attribute format")
	}

	return &att, nil
}

func (d *Decoder) parseURI(value string) (string, error) {
	if d.s.MediaDescs != nil {
		return "", fmt.Errorf("URI must be specified before the first media field")
	}

	if d.s.URI != "" {
		return "", fmt.Errorf("Multiple URIs")
	}

	return value, nil
}

func (d *Decoder) parseEmail(value string) (string, error) {
	if d.s.MediaDescs != nil {
		return "", fmt.Errorf("Email must be specified before the first media field")
	}

	return value, nil
}

func (d *Decoder) parsePhoneNumber(value string) (string, error) {
	if d.s.MediaDescs != nil {
		return "", fmt.Errorf("Phone number must be specified before the first media field")
	}

	return value, nil
}

func (d *Decoder) parseOriginator(value string) (*Origin, error) {
	var err error

	fields := strings.Split(value, " ")
	if len(fields) != 6 {
		return nil, fmt.Errorf("Wrong originator format")
	}

	var origin Origin
	origin.Username = fields[0]
	origin.SessID, err = strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Wrong originator.sess-id format")
	}
	origin.SessVersion, err = strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Wrong originator.sess-version format")
	}
	origin.Nettype = fields[3]
	origin.Addrtype = fields[4]
	origin.UnicastAddress = fields[5]

	return &origin, nil
}

func (d *Decoder) parseConnection(value string) (*Connection, error) {
	var err error

	fields := strings.Split(value, " ")
	if len(fields) != 3 {
		return nil, fmt.Errorf("Wrong connection format")
	}

	var connection Connection
	connection.Nettype = fields[0]
	connection.Addrtype = fields[1]

	fields = strings.Split(fields[2], "/")
	connection.ConnectionAddr = fields[0]

	if connection.Addrtype == "IP4" {
		if len(fields) > 1 {
			connection.TTL, err = strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("Wrong connection.TTL format")
			}
		}
		if len(fields) > 2 {
			connection.AddressesNum, err = strconv.ParseInt(fields[2], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("Wrong connection.addresses-num format")
			}
		} else {
			connection.AddressesNum = 1
		}
	} else if connection.Addrtype == "IP6" {
		if len(fields) > 1 {
			connection.AddressesNum, err = strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("Wrong connection.addresses-num format")
			}
		} else {
			connection.AddressesNum = 1
		}
	}

	return &connection, nil
}

func (d *Decoder) parseBandwidth(value string) (*Bandwidth, error) {
	var err error
	var bandwidth Bandwidth

	fields := strings.Split(value, ":")
	if len(fields) != 2 {
		return nil, fmt.Errorf("Wrong bandwidth format")
	}

	bandwidth.Type = fields[0]
	bandwidth.Value, err = strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("Wrong bandwidth format")
	}

	return &bandwidth, nil
}

func (d *Decoder) parseTiming(value string) (*Timing, error) {
	var timing Timing
	var err error

	fields := strings.Split(value, " ")
	if len(fields) != 2 {
		return nil, fmt.Errorf("Wrong timing format")
	}

	timing.Start, err = strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Wrong timing format")
	}

	timing.Stop, err = strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Wrong timing format")
	}

	return &timing, err
}

func (d *Decoder) parseTimeZones(value string) ([]*TimeZone, error) {
	var timeZones []*TimeZone
	var err error

	fields := strings.Split(value, " ")

	if len(fields)%2 != 0 {
		return nil, fmt.Errorf("Wrong time zone format")
	}

	for i := 0; i < len(fields); i += 2 {
		var timeZone TimeZone

		timeZone.Time, err = strconv.ParseInt(fields[i], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Error while parsing time zone: %v", err)
		}

		timeZone.Offset, err = d.parseTime(fields[i+1])
		if err != nil {
			return nil, fmt.Errorf("Error while parsing time zone: %v", err)
		}

		timeZones = append(timeZones, &timeZone)
	}

	return timeZones, nil
}

func (d *Decoder) parseSessionName(value string) (string, error) {
	if d.s.SessionName != "" {
		return "", fmt.Errorf("Multiple session names")
	}
	if value == "" {
		return "", fmt.Errorf("Session name must not be empty")
	}
	return value, nil
}

func (d *Decoder) parseRepeatTime(value string) (*RepeatTime, error) {
	var err error
	var repeat RepeatTime

	fields := strings.Split(value, " ")
	if len(fields) < 3 {
		return nil, fmt.Errorf("Wrong repeat time format")
	}

	repeat.Interval, err = d.parseTime(fields[0])
	if err != nil {
		return nil, fmt.Errorf("Error while parsing time: %v", err)
	}

	repeat.Duration, err = d.parseTime(fields[1])
	if err != nil {
		return nil, fmt.Errorf("Error while parsing time: %v", err)
	}

	for i := 2; i < len(fields); i += 1 {
		offset, err := d.parseTime(fields[i])
		if err != nil {
			return nil, fmt.Errorf("Error while parsing time: %v", err)
		}
		repeat.Offsets = append(repeat.Offsets, offset)
	}

	return &repeat, nil
}

func (d *Decoder) parseMedia(value string) (string, error) {
	if !inSet(value, []string{"audio", "video", "text", "application", "message"}) {
		return "", fmt.Errorf("Wrong media: %v", value)
	}
	return value, nil
}

func (d *Decoder) parsePort(value string) (int64, error) {
	port, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		return 0, fmt.Errorf("Error while parsing port: %v", err)
	}

	if port < 0 || port > 65536 {
		return 0, fmt.Errorf("Error while parsing poert: port out of range")
	}

	return port, nil
}

func (d *Decoder) parsePortsNum(value string) (int64, error) {
	portsNum, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		return 0, fmt.Errorf("Error while parsing ports num: %v", err)
	}

	return portsNum, nil
}

func inSet(key string, values []string) bool {
	for _, value := range values {
		if value == key {
			return true
		}
	}
	return false
}

func (d *Decoder) parseMediaDesc(line string, lineNum int) (*MediaDesc, error) {
	var mediaDesc MediaDesc
	var err error

	fields := strings.Split(line, " ")

	mediaDesc.Media, err = d.parseMedia(fields[0])
	if err != nil {
		return nil, fmt.Errorf("Wrong media discription format: %v", err)
	}

	parts := strings.Split(fields[1], "/")
	mediaDesc.Port, err = d.parsePort(parts[0])
	if err != nil {
		return nil, fmt.Errorf("Wrong media discription format: %v", err)
	}

	if len(parts) > 1 {
		mediaDesc.PortsNum, err = d.parsePortsNum(parts[1])

		if err != nil {
			fmt.Errorf("Wrong media discription format: %v", err)
		}
	} else {
		mediaDesc.PortsNum = 1
	}

	if len(parts) > 2 {
		return nil, fmt.Errorf("Wrong media discription format")
	}

	for _, proto := range strings.Split(fields[2], "/") {
		if !inSet(proto, []string{"UDP", "RTP", "AVP", "SAVP", "SAVPF", "TLS", "DTLS", "SCTP", "AVPF", "TCP", "MSRP"}) {
			return nil, fmt.Errorf("Wrong media discription format: wrong protocol format")
		}
		mediaDesc.Proto = append(mediaDesc.Proto, proto)
	}

	fields = fields[3:]
	for _, field := range fields {
		mediaDesc.Fmts = append(mediaDesc.Fmts, field)
	}

	return &mediaDesc, nil
}

func (d *Decoder) parseInforamtion(line string) string {
	return line
}

func (d *Decoder) parseMediaLine(line string, lineNum int) error {
	var err error
	media := d.s.MediaDescs[len(d.s.MediaDescs)-1]

	if (len(line) < 2) || (line[1] != '=') {
		return fmt.Errorf("Wrong line format, line %v", lineNum)
	}

	key, value := line[0], line[2:]
	switch key {
	case 'i':
		if media.Information != "" {
			err = fmt.Errorf("Two information per media")
		} else {
			media.Information = d.parseInforamtion(value)
		}
	case 'c':
		connectionData, connErr := d.parseConnection(value)
		if connErr != nil {
			err = connErr
		} else {
			media.Connections = append(media.Connections, connectionData)
		}
	case 'b':
		bandwidth, err := d.parseBandwidth(value)
		if err == nil {
			media.Bandwidths = append(media.Bandwidths, bandwidth)
		}
	case 'k':
		key, keyErr := d.parseEncryptionKey(value)
		if keyErr != nil {
			err = keyErr
		} else {
			media.EncryptionKeys = append(media.EncryptionKeys, key)
		}
	case 'a':
		attribute, attErr := d.parseAttribute(value)
		if attErr != nil {
			err = attErr
		} else {
			d.s.MediaDescs[len(d.s.MediaDescs)-1].Attributes = append(d.s.MediaDescs[len(d.s.MediaDescs)-1].Attributes, attribute)
		}
	default:
		return fmt.Errorf("Unknown parameter type, line %v", lineNum)
	}
	return err
}

func (d *Decoder) parseSessionLine(line string, lineNum int) error {
	var err error

	if (len(line) < 2) || (line[1] != '=') {
		return fmt.Errorf("Wrong line format, line %v", lineNum)
	}

	key, value := line[0], line[2:]
	switch key {
	case 'i':
		if d.s.Information != "" {
			err = fmt.Errorf("Two information per media")
		} else {
			d.s.Information = d.parseInforamtion(value)
		}
	case 'v':
		d.s.Version, err = d.parseVersion(value)
	case 'o':
		d.s.Originator, err = d.parseOriginator(value)
	case 's':
		d.s.SessionName, err = d.parseSessionName(value)
	case 'u':
		d.s.URI, err = d.parseURI(value)
	case 'e':
		email, valErr := d.parseEmail(value)
		if valErr != nil {
			err = valErr
		} else {
			d.s.Emails = append(d.s.Emails, email)
		}
	case 'p':
		phone, valErr := d.parsePhoneNumber(value)
		if valErr != nil {
			err = valErr
		} else {
			d.s.PhoneNumbers = append(d.s.PhoneNumbers, phone)
		}
	case 'c':
		if d.s.ConnectionData != nil {
			err = fmt.Errorf("Multiple connection data descriptions per session")
		} else {
			d.s.ConnectionData, err = d.parseConnection(value)
		}
	case 'b':
		bandwidth, bandErr := d.parseBandwidth(value)
		if bandErr != nil {
			return err
		}
		d.s.Bandwidth = append(d.s.Bandwidth, bandwidth)

	case 'z':
		timeZones, tzErr := d.parseTimeZones(value)
		if tzErr != nil {
			err = tzErr
		} else {
			d.s.TimeZones = append(d.s.TimeZones, timeZones...)
		}
	case 'k':
		key, keyErr := d.parseEncryptionKey(value)
		if keyErr != nil {
			err = keyErr
		} else {
			d.s.EncryptionKeys = append(d.s.EncryptionKeys, key)
		}
	case 'a':
		attribute, attErr := d.parseAttribute(value)
		if attErr != nil {
			err = attErr
		} else {
			d.s.Attributes = append(d.s.Attributes, attribute)
		}
	case 't':
		timing, timingErr := d.parseTiming(value)
		if timingErr != nil {
			return err
		} else {
			d.s.Timings = append(d.s.Timings, timing)
		}
	case 'r':
		repeat, repErr := d.parseRepeatTime(value)
		if repErr != nil {
			err = repErr
		} else {
			if len(d.s.Timings) == 0 {
				err = fmt.Errorf("r= should not be specified before t=")
			} else {
				d.s.Timings[len(d.s.Timings)-1].RepeatTimes = append(d.s.Timings[len(d.s.Timings)-1].RepeatTimes, repeat)
			}
		}
	default:
		return fmt.Errorf("Unknown parameter type, line %v", lineNum)
	}

	return err
}

func (d *Decoder) Decode() (*Session, error) {
	var err error

	scanner := bufio.NewScanner(d.r)
	lineNum := 1

	for scanner.Scan() {
		line := scanner.Text()

		if line[0] == 'm' {
			mediaDesc, mediaErr := d.parseMediaDesc(line[2:], lineNum)
			if mediaErr != nil {
				err = mediaErr
			} else {
				d.s.MediaDescs = append(d.s.MediaDescs, mediaDesc)
			}
		} else if d.s.MediaDescs != nil {
			err = d.parseMediaLine(line, lineNum)
		} else {
			err = d.parseSessionLine(line, lineNum)
		}
		if err != nil {
			return nil, fmt.Errorf("Error while parsing: %v", err)
		}

		lineNum += 1
	}

	if scanner.Err() != nil {
		return nil, fmt.Errorf("Error while reading from reader: %v", scanner.Err())
	}

	// TODO: check connection

	return d.s, nil
}
