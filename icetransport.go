package webrtc

type RTCIceTransport struct {
	role           *RTCIceRole
	component      *RTCIceComponent
	state          *RTCIceTransportState
	gatheringState *RTCIceGatheringState
}
