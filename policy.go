package webrtc

type RTCIceTransportPolicy string

const (
	RTCIceTransportPolicyRelay = "relay"
	RTCIceTransportPolicyAll   = "all"
)

type RTCBundlePolicy string

const (
	RTCBundlePolicyBalancded = "balancded"
	RTCBundlePolicyMaxCompat = "max-compat"
	RTCBundlePolicyMaxBundle = "max-bundle"
)

type RTCRtcpMuxPolicy string

const (
	RTCRtcpMuxPolicyRequire = "require"
)
