package webrtc

type RTCIceGathererState string

const (
	RTCIceGathererStateNew       = "new"
	RTCIceGathererStateGathering = "gathering"
	RTCIceGathererStateComplete  = "complete"
)
