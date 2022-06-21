package webrtc

type RTCIceConnectionState string

const (
	RTCIceConnectionStateClosed       = "closed"
	RTCIceConnectionStateFailed       = "failed"
	RTCIceConnectionStateDisconnected = "disconnected"
	RTCIceConnectionStateNew          = "new"
	RTCIceConnectionStateChecking     = "checking"
	RTCIceConnectionStateCompleted    = "completed"
	RTCIceConnectionStateConnected    = "connected"
)
