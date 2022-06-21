package webrtc

type RTCDtlsTransport struct {
	iceTransport *RTCIceTransport
	state        *RTCDtlsTransportState
}
