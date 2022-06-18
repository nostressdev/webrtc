package webrtc

type RTPCodecParameters struct {
	MimeType    string
	PayloadType uint8

	ClockRate   uint32
	Channels    uint16
	SDPFmtpLine string
}
