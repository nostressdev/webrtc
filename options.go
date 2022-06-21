package webrtc

type RTCOfferAnswerOptions struct {
}

type RTCAnswerOptions struct {
	RTCOfferAnswerOptions
}

type RTCOfferOptions struct {
	RTCOfferAnswerOptions
	ICERestart bool
}
