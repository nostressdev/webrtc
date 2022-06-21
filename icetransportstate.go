package webrtc

type RTCIceTransportState string

const (
	RTCIceTransportStateNew          = "new"
	RTCIceTransportStateChecking     = "checking"
	RTCIceTransportStateConnected    = "connected"
	RTCIceTransportStateCompleted    = "completed"
	RTCIceTransportStateDisconnected = "disconnected"
	RTCIceTransportStateFailed       = "failed"
	RTCIceTransportStateClosed       = "closed"
)
