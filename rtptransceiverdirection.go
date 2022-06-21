package webrtc

type RTCRtpTransceiverDirection string

const (
	RTCRtpTransceiverDirectionSendrecv = "sendrecv"
	RTCRtpTransceiverDirectionSendonly = "sendonly"
	RTCRtpTransceiverDirectionRecvonly = "recvonly"
	RTCRtpTransceiverDirectionInactive = "inactive"
	RTCRtpTransceiverDirectionStopped  = "stopped"
)
