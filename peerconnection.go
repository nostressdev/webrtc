package webrtc

type RTCPeerConnection struct {
	configuration            RTCConfiguration
	localDescription         RTCSessionDescription
	currentLocalDescription  RTCSessionDescription
	pendingLocalDescription  RTCSessionDescription
	remoteDescription        RTCSessionDescription
	currentRemoteDescription RTCSessionDescription
	pendingRemoteDescription RTCSessionDescription
	signalingState           RTCSignalingState
	iceGatheringState        RTCIceGatheringState
	iceConnectionState       RTCIceConnectionState
	connectionState          RTCPeerConnectionState
	canTrickleIceCandidates  bool

	// TODO(@alisa-vernigor): add event handlers
}
