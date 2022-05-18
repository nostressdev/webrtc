package webrtc

type RTCRtpSender struct {
	track     *MediaStreamTrack
	transport *RTCDtlsTransport
}
