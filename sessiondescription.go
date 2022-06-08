package webrtc

import (
	"bytes"

	"github.com/nostressdev/webrtc/sdp"
)

type RTCSessionDescription struct {
	SDPType   RTCSdpType
	SDPString string
	Session   *sdp.Session
}

func (s *RTCSessionDescription) setSession(session *sdp.Session) {
	s.Session = session
	var res string
	buf := bytes.NewBufferString((res))

	sdp.NewEncoder(buf).Encode(session)
	s.SDPString = buf.String()
}
