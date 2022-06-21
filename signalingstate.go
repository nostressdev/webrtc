package webrtc

type RTCSignalingState string

const (
	RTCSignalingStateStable             = "stable"
	RTCSignalingStateHaveLocalOffer     = "have-local-offer"
	RTCSignalingStateHaveRemoteOffer    = "have-remote-offer"
	RTCSignalingStateHaveLocalPranswer  = "have-local-pranswer"
	RTCSignalingStateHaveRemotePranswer = "have-remote-pranswer"
	RTCSignalingStateClosed             = "closed"
)
