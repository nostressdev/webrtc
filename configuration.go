package webrtc

type RTCConfiguration struct {
	IceServers           []RTCIceServer
	IceTransportPolicy   RTCIceTransportPolicy
	BundlePolicy         RTCBundlePolicy
	RtcpMuxPolicy        RTCRtcpMuxPolicy
	Certificates         []RTCCertificate
	IceCandidatePoolSize uint8
}
