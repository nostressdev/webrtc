package webrtc

type RTCPeerConnectionState string

const (
	RTCPeerConnectionStateClosed       = "closed"
	RTCPeerConnectionStateFailed       = "failed"
	RTCPeerConnectionStateDisconnected = "disconnected"
	RTCPeerConnectionStateNew          = "new"
	RTCPeerConnectionStateConnecting   = "connecting"
	RTCPeerConnectionStateConnected    = "connected"
)
