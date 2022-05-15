package webrtc

type RTCSdpType string

const (
	RTCSdpTypeOffer    = "offer"
	RTCSdpTypePranswer = "pranswer"
	RTCSdpTypeAnswer   = "answer"
	RTCSdpTypeRollback = "rollback"
)
