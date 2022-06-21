package webrtc

type MediaStreamTrack struct {
	kind    string
	id      string
	label   string
	enabled bool
	muted   bool

	// TODO(@alisa-vernigor): add handlers

	readyState MediaStreamTrackState
}
