package webrtc

type RTCIceServer struct {
	URLs           []string
	Username       string
	Credential     string
	CredentialType RTCIceCredentialType
}

type RTCIceCredentialType string

const (
	RTCIceCredentialTypePassword RTCIceCredentialType = "password"
)
