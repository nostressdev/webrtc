package webrtc

type RTCDtlsTransportState string

const (
	RTCDtlsTransportStateNew        = "new"
	RTCDtlsTransportStateConnecting = "connecting"
	RTCDtlsTransportStateConnected  = "connected"
	RTCDtlsTransportStateClosed     = "closed"
	RTCDtlsTransportStateFailed     = "failed"
)
