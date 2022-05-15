package webrtc

type RTCIceGatheringState string

const (
	RTCIceGatheringStateNew       = "new"
	RTCIceGatheringStateGathering = "gathering"
	RTCIceGatheringStateComplete  = "complete"
)
