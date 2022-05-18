package webrtc

type RTCRtpReceiver struct {
	track     *MediaStreamTrack
	transport *RTCDtlsTransport
}
