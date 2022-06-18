package webrtc

type RTCRtpTransceiver struct {
	mid              string
	sender           *RTCRtpSender
	receiver         *RTCRtpReceiver
	direction        RTCRtpTransceiverDirection
	currentDirection RTCRtpTransceiverDirection
	stopped          bool
	stopping         bool

	codecs []RTPCodecParameters
}
