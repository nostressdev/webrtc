package webrtc

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nostressdev/webrtc/sdp"
)

type RTCPeerConnection struct {
	sessVersion                             int64
	certsGenerated                          *sync.Cond
	configuration                           *RTCConfiguration
	localDescription                        *RTCSessionDescription
	currentLocalDescription                 *RTCSessionDescription
	pendingLocalDescription                 *RTCSessionDescription
	remoteDescription                       *RTCSessionDescription
	currentRemoteDescription                *RTCSessionDescription
	pendingRemoteDescription                *RTCSessionDescription
	signalingState                          RTCSignalingState
	iceGatheringState                       *RTCIceGatheringState
	iceConnectionState                      *RTCIceConnectionState
	connectionState                         *RTCPeerConnectionState
	canTrickleIceCandidates                 bool
	isClosed                                bool
	documentOrigin                          *sdp.Origin
	negotiationNeeded                       bool
	updateNegotiationNeededFlagOnEmptyChain bool
	lastCreatedOffer                        string
	lastCreatedAnswer                       string
	rtpTransceivers                         []*RTCRtpTransceiver
	midCounter                              int64
	// TODO(@alisa-vernigor): sctpTransport, earlyCandidates, operations, localIceCredentialsToReplace
	// TODO(@alisa-vernigor): add event handlers
}

type MediaType string

const (
	MediaTypeAudio       = "audio"
	MediaTypeVideo       = "video"
	MediaTypeText        = "text"
	MediaTypeApplication = "application"
	MediaTypeMessage     = "message"
)

func (pc *RTCPeerConnection) matchIceOptions(session *sdp.Session, sessionToMatch *sdp.Session) {
	iceOptionsValues := sessionToMatch.GetAttribute("ice-options")
	for _, iceOptionsValue := range iceOptionsValues {
		if iceOptionsValue == "trickle ice2" || iceOptionsValue != "ice2 trickle" {
			return
		}
	}

	session.ExcludeAttribute("ice-options")
}

func (pc *RTCPeerConnection) findTransceiverByMid(mid string) *RTCRtpTransceiver {
	for _, tranceiver := range pc.rtpTransceivers {
		if tranceiver.mid == mid {
			return tranceiver
		}
	}
	return nil
}

func (pc *RTCPeerConnection) addLSGroups(session *sdp.Session, sessionToMatch *sdp.Session) {
	groups := sessionToMatch.GetAttribute("group")
	var LSGroups [][]string

	for _, group := range groups {
		if len(group) > 2 && group[:1] == "LS" {
			LSGroups = append(LSGroups, strings.Split(group, " ")[1:])
		}
	}

	for _, LSGroup := range LSGroups {
		msidsToMidStrings := make(map[string][]string)
		for _, mid := range LSGroup {
			tranceiver := pc.findTransceiverByMid(mid)
			if tranceiver == nil {
				continue
			}
			for _, mediaStreamId := range tranceiver.sender.associatedMediaStreamIds {
				msidsToMidStrings[mediaStreamId] = append(msidsToMidStrings[mediaStreamId], tranceiver.mid)
			}

			if len(tranceiver.sender.associatedMediaStreamIds) == 0 {
				for msid, _ := range msidsToMidStrings {
					msidsToMidStrings[msid] = append(msidsToMidStrings[msid], tranceiver.mid)
				}
			}
		}

		for msid, _ := range msidsToMidStrings {
			if len(msidsToMidStrings[msid]) > 1 {
				session.AddAttribute("group", "LS"+" "+strings.Join(msidsToMidStrings[msid], " "))
			}
		}
	}
}

func (pc *RTCPeerConnection) getCodecNameByFmt(media *sdp.MediaDesc, format string) (string, error) {
	rtpmaps := media.GetAttribute("rtpmap")
	for _, rtpmap := range rtpmaps {
		parsedRtpmap := strings.Split(rtpmap, " ")
		if len(parsedRtpmap) < 2 {
			return "", fmt.Errorf("wrong rtpmap format")
		}
		if parsedRtpmap[0] == format {
			return strings.Split(strings.Join(parsedRtpmap[1:], " "), "/")[0], nil
		}
	}
	return "", nil
}

func (pc *RTCPeerConnection) containSupportedCodecs(media *sdp.MediaDesc, supportedCodecs []string) (bool, error) {
	supportedCodecsMap := make(map[string]bool)
	for _, codec := range supportedCodecs {
		supportedCodecsMap[codec] = true
	}

	for _, format := range media.Fmts {
		codecName, err := pc.getCodecNameByFmt(media, format)
		if err != nil {
			return false, fmt.Errorf("error while getting codec name: %v", err)
		}

		if supportedCodecsMap[codecName] {
			return true, nil
		}
	}
	return false, nil
}

type iceCredentials struct {
	iceUfrag string
	icePwd   string
}

func (pc *RTCPeerConnection) genereateIceCredentials() *iceCredentials {
	return &iceCredentials{
		iceUfrag: randStringWithCharset(4, alphaNumCharset),
		icePwd:   randStringWithCharset(22, alphaNumCharset),
	}
}

func (pc *RTCPeerConnection) addMatchedDirection(mediaDesc *sdp.MediaDesc, directionToMatch RTCRtpTransceiverDirection) {
	if directionToMatch == RTCRtpTransceiverDirectionRecvonly {
		mediaDesc.AddAttribute(RTCRtpTransceiverDirectionSendonly, " ")
	} else if directionToMatch == RTCRtpTransceiverDirectionSendonly {
		mediaDesc.AddAttribute(RTCRtpTransceiverDirectionRecvonly, " ")
	} else if directionToMatch == RTCRtpTransceiverDirectionSendrecv {
		mediaDesc.AddAttribute(RTCRtpTransceiverDirectionSendrecv, " ")
	} else {
		mediaDesc.AddAttribute(string(directionToMatch), " ")
	}
}

func (pc *RTCPeerConnection) getDirection(mediaDesc *sdp.MediaDesc) RTCRtpTransceiverDirection {
	if mediaDesc.GetAttribute(RTCRtpTransceiverDirectionSendrecv) != nil {
		return RTCRtpTransceiverDirectionSendrecv
	}
	if mediaDesc.GetAttribute(RTCRtpTransceiverDirectionSendonly) != nil {
		return RTCRtpTransceiverDirectionSendonly
	}
	if mediaDesc.GetAttribute(RTCRtpTransceiverDirectionRecvonly) != nil {
		return RTCRtpTransceiverDirectionRecvonly
	}
	if mediaDesc.GetAttribute(RTCRtpTransceiverDirectionInactive) != nil {
		return RTCRtpTransceiverDirectionInactive
	}
	return RTCRtpTransceiverDirectionStopped
}

func (pc *RTCPeerConnection) addFingerprints(mediaDesc *sdp.MediaDesc, fingerprints []*RTCDtlsFingerprint) {
	for _, fingerprint := range fingerprints {
		mediaDesc.AddAttribute("fingerprint", fingerprint.Algorithm+" "+fingerprint.Value)
	}
}

type setupType string

const (
	setupTypeActpass = "actpass"
	setupTypeActive  = "active"
	setupTypePassive = "passive"
)

func (pc *RTCPeerConnection) addMatchedSetup(mediaDesc *sdp.MediaDesc, mediaDescToMatch *sdp.MediaDesc) error {
	setupToMatch := mediaDescToMatch.GetAttribute("setup")
	if setupToMatch == nil {
		return fmt.Errorf("no setup in offer media description")
	}

	if setupToMatch[0] == setupTypeActpass || setupToMatch[0] == setupTypePassive {
		mediaDesc.AddAttribute("setup", setupTypeActive)
	} else {
		mediaDesc.AddAttribute("setup", setupTypePassive)
	}

	return nil
}

func (pc *RTCPeerConnection) addIceCredentials(mediaDesc *sdp.MediaDesc, credentials *iceCredentials) {
	mediaDesc.AddAttribute("ice-ufrag", credentials.iceUfrag)
	mediaDesc.AddAttribute("ice-pwd", credentials.icePwd)
}

func (pc *RTCPeerConnection) addMatchedIceCredentials(mediaDesc *sdp.MediaDesc, mediaDescToMatch *sdp.MediaDesc) error {
	iceUfrag := mediaDescToMatch.GetAttribute("ice-ufrag")
	if iceUfrag == nil {
		return fmt.Errorf("error getting ice credentials")
	}
	icePwd := mediaDescToMatch.GetAttribute("ice-pwd")
	if icePwd == nil {
		return fmt.Errorf("error getting ice credentials")
	}

	pc.addIceCredentials(mediaDesc, &iceCredentials{iceUfrag[0], icePwd[0]})

	return nil
}

func (pc *RTCPeerConnection) generateMediaSectionBase(isBundled bool, fingerprints []*RTCDtlsFingerprint,
	kind MediaType, tlsID string) *sdp.MediaDesc {
	mediaDesc := &sdp.MediaDesc{}
	mediaDesc.Connections = []*sdp.Connection{
		{
			Nettype:        sdp.NetworkInternet,
			Addrtype:       sdp.TypeIPv4,
			ConnectionAddr: "0.0.0.0",
			AddressesNum:   1,
		},
	}

	if isBundled {
		mediaDesc.Port = 0
	} else {
		mediaDesc.Port = 9
		pc.addFingerprints(mediaDesc, fingerprints)
		mediaDesc.AddAttribute("tls-id", tlsID)
	}
	mediaDesc.Media = string(kind)
	return mediaDesc
}

func (pc *RTCPeerConnection) addCodecs(mediaDesc *sdp.MediaDesc, kind MediaType) {
	if kind == MediaTypeAudio {
		pc.addAudioCodecs(mediaDesc)
	} else if kind == MediaTypeVideo {
		pc.addVideoCodecs(mediaDesc)
	}
}

func (pc *RTCPeerConnection) generateUnmatchedMediaSectionBase(fingerprints []*RTCDtlsFingerprint,
	isBundled bool, tranceiver *RTCRtpTransceiver, tlsID string,
	mids []string, kind MediaType) *sdp.MediaDesc {

	mediaDesc := pc.generateMediaSectionBase(isBundled, fingerprints, kind, tlsID)

	pc.addCodecs(mediaDesc, kind)

	if isBundled {
		mediaDesc.AddAttribute("bundle-only", " ")
	} else {
		mediaDesc.AddAttribute("setup", "actpass")

		pc.addIceCredentials(mediaDesc, pc.genereateIceCredentials())
	}
	mediaDesc.Proto = []string{"UDP", "TLS", "RTP", "SAVPF"}

	mid := strconv.FormatInt(pc.midCounter, 10)
	if tranceiver != nil {
		tranceiver.mid = mid
	}
	mediaDesc.AddAttribute("mid", mid)
	mids = append(mids, mid)
	pc.midCounter++

	return mediaDesc
}

func (pc *RTCPeerConnection) generateMatchedMediaSectionBase(fingerprints []*RTCDtlsFingerprint,
	kind MediaType, tlsID string, isBundled bool,
	mediaDescToMatch *sdp.MediaDesc) (*sdp.MediaDesc, error) {
	var mediaDesc *sdp.MediaDesc

	mediaDesc.Media = string(kind)

	var credentials *iceCredentials
	if isBundled {
		iceUfrag := mediaDescToMatch.GetAttribute("ice-ufrag")
		if iceUfrag == nil {
			return nil, fmt.Errorf("error getting ice credentials")
		}
		icePwd := mediaDescToMatch.GetAttribute("ice-pwd")
		if icePwd == nil {
			return nil, fmt.Errorf("error getting ice credentials")
		}
		credentials = &iceCredentials{iceUfrag: iceUfrag[0], icePwd: icePwd[0]}
	} else {
		credentials = pc.genereateIceCredentials()
		err := pc.addMatchedSetup(mediaDesc, mediaDescToMatch)
		if err != nil {
			return nil, fmt.Errorf("error while adding setup attribute: %v", err)
		}

		mediaDesc.AddAttribute("tls-id", tlsID)
		pc.addFingerprints(mediaDesc, fingerprints)
	}
	mediaDesc.Proto = mediaDescToMatch.Proto

	pc.addIceCredentials(mediaDesc, credentials)

	if mid := mediaDescToMatch.GetAttribute("mid"); mid != nil {
		mediaDesc.AddAttribute("mid", mid[0])
	}
	return mediaDesc, nil
}

func (pc *RTCPeerConnection) addMatchedMediaSection(session *sdp.Session, mediaDescToMatch *sdp.MediaDesc, kind MediaType, tlsID string, isBundled bool, fingerprints []*RTCDtlsFingerprint, tranceiver *RTCRtpTransceiver) error {
	mediaDesc, err := pc.generateMatchedMediaSectionBase(fingerprints, kind, tlsID, isBundled, mediaDescToMatch)
	if err != nil {
		return fmt.Errorf("error while generating media section base: %v")
	}

	pc.addMatchedDirection(mediaDesc, pc.getDirection(mediaDescToMatch))
	for _, mediaStreamId := range tranceiver.sender.associatedMediaStreamIds {
		mediaDesc.AddAttribute("msid", mediaStreamId)
	}

	if mediaDescToMatch.GetAttribute("rtcp-rsize") != nil {
		mediaDesc.AddAttribute("rtcp-rsize", " ")
	}

	session.MediaDescs = append(session.MediaDescs, mediaDesc)

	return nil
}

func (pc *RTCPeerConnection) addMatchedDataSection(session *sdp.Session, mediaDescToMatch *sdp.MediaDesc, fingerprints []*RTCDtlsFingerprint, tlsID string, supportsBundle bool) error {
	mediaDesc, err := pc.generateMatchedMediaSectionBase(fingerprints, MediaTypeApplication, tlsID, false, mediaDescToMatch)
	if err != nil {
		return fmt.Errorf("error while generating mediaDesc: %v", err)
	}

	mediaDesc.Proto = mediaDescToMatch.Proto
	mediaDesc.AddAttribute("sctp-port", "5000")

	session.MediaDescs = append(session.MediaDescs, mediaDesc)

	return nil
}

func (pc *RTCPeerConnection) getRejectedVector(session *sdp.Session, sessionToMatch *sdp.Session, supportsBundle *bool) ([]bool, error) {
	var isRejected []bool

	bundlePolicy := pc.configuration.BundlePolicy
	isFirstInGroup := map[MediaType]bool{
		MediaTypeAudio:       true,
		MediaTypeVideo:       true,
		MediaTypeText:        true,
		MediaTypeApplication: true,
		MediaTypeMessage:     true,
	}
	isBundled := make(map[string]bool)

	groups := sessionToMatch.GetAttribute("group")
	for _, group := range groups {
		if len(group) > 6 && group[:6] == "BUNDLE" {
			for _, mid := range strings.Split(group, " ")[1:] {
				isBundled[mid] = true
			}
			*supportsBundle = true
		}
	}

	for _, mediaDesc := range sessionToMatch.MediaDescs {
		midAttr := mediaDesc.GetAttribute("mid")
		if midAttr == nil {
			isRejected = append(isRejected, true)
			continue
		}
		mid := midAttr[0]

		if mediaDesc.Media == MediaTypeApplication {
			isRejected = append(isRejected, false)
			continue
		}

		if mediaDesc.Media == MediaTypeAudio {
			gotSupportedCodecs, err := pc.containSupportedCodecs(mediaDesc, []string{"opus"})
			if err != nil {
				return nil, fmt.Errorf("error with codecs: %v", err)
			}
			if !gotSupportedCodecs {
				isRejected = append(isRejected, true)
				continue
			}
		}

		if mediaDesc.Media == MediaTypeVideo {
			gotSupportedCodecs, err := pc.containSupportedCodecs(mediaDesc, []string{"H264 AVC"})
			if err != nil {
				return nil, fmt.Errorf("error with codecs: %v", err)
			}
			if !gotSupportedCodecs {
				isRejected = append(isRejected, true)
				continue
			}
		}

		if mediaDesc.Media != MediaTypeVideo && mediaDesc.Media != MediaTypeAudio && mediaDesc.Media != MediaTypeApplication {
			isRejected = append(isRejected, true)
			continue
		}

		tranceiver := pc.findTransceiverByMid(mid)
		if tranceiver.stopped {
			isRejected = append(isRejected, true)
			continue
		}

		if pc.isBundleOnly(bundlePolicy, isFirstInGroup, MediaType(tranceiver.sender.track.kind)) {
			isRejected = append(isRejected, true)
			continue
		}

		if isBundled[mid] && mediaDesc.Port == 0 {
			isRejected = append(isRejected, true)
			continue
		}

		isRejected = append(isRejected, false)
	}

	return isRejected, nil
}

func (pc *RTCPeerConnection) generateFingerprints() ([]*RTCDtlsFingerprint, error) {
	var err error
	var fingerprints []*RTCDtlsFingerprint
	for _, cert := range pc.configuration.Certificates {
		fingerprints, err = cert.getFingerprints()
		if err != nil {
			return nil, fmt.Errorf("error while getting fingerprints: %v", err)
		}
	}
	return fingerprints, nil
}

func (pc *RTCPeerConnection) addRejectedMediaSection(session *sdp.Session, kind MediaType) {
	mediaDesc := &sdp.MediaDesc{
		Media:    string(kind),
		Port:     0,
		Proto:    []string{"UDP", "TLS", "RTP", "SAVPF"},
		Fmts:     []string{""},
		PortsNum: 1,
	}
	session.MediaDescs = append(session.MediaDescs, mediaDesc)
}

func (pc *RTCPeerConnection) addMatchedMediaSections(session *sdp.Session, sessionToMatch *sdp.Session) error {
	supportsBundle := false

	isRejected, err := pc.getRejectedVector(session, sessionToMatch, &supportsBundle)
	if err != nil {
		fmt.Errorf("error while analyzing rejected media sections: %v", err)
	}

	fingerprints, err := pc.generateFingerprints()
	if err != nil {
		return fmt.Errorf("error while generating fingerprints: %v", err)
	}

	tlsID := randStringWithCharset(120, alphaNumCharset)

	for index, mediaDescToMatch := range sessionToMatch.MediaDescs {
		var err error
		if isRejected[index] {
			pc.addRejectedMediaSection(session, MediaType(mediaDescToMatch.Media))
		} else {
			if mediaDescToMatch.Media == MediaTypeApplication {
				err = pc.addMatchedDataSection(session, mediaDescToMatch, fingerprints, tlsID, true)
			} else {
				err = pc.addMatchedMediaSection(session, mediaDescToMatch, MediaType(mediaDescToMatch.Media),
					tlsID, supportsBundle && (mediaDescToMatch.GetAttribute("bundle-only") != nil),
					fingerprints, pc.findTransceiverByMid(mediaDescToMatch.GetAttribute("mid")[0]))
			}
		}

		if err != nil {
			return fmt.Errorf("error while generating media sections: %v", err)
		}
	}

	return nil
}

func (pc *RTCPeerConnection) createAnswer(answer *RTCSessionDescription) (*sdp.Session, error) {
	session := &sdp.Session{}

	pc.addMatchedSessionLevelAttributes(session)
	pc.addLSGroups(session, pc.currentLocalDescription.Session)
	pc.addMatchedMediaSections(session, pc.currentRemoteDescription.Session)

	return session, nil
}

func (pc *RTCPeerConnection) processLocalDescription(description *RTCSessionDescription) error {
	switch description.SDPType {
	case RTCSdpTypeRollback:
		// do rollback
	case RTCSdpTypeOffer:
		if pc.signalingState != RTCSignalingStateStable && pc.signalingState != RTCSignalingStateHaveLocalOffer {
			return fmt.Errorf("description type offer, but signaling state is not stable/have-local-offer")
		}

		if description.SDPString != pc.lastCreatedOffer {
			return fmt.Errorf("description was altered since last call to createOffer")
		}
	case RTCSdpTypePranswer, RTCSdpTypeAnswer:
		if pc.signalingState != RTCSignalingStateHaveRemoteOffer && pc.signalingState != RTCSignalingStateHaveLocalPranswer {
			return fmt.Errorf("description type answer/pranswer, but signaling state is not have-remote-offer/have-local-pranswer")
		}

		if description.SDPString != pc.lastCreatedAnswer {
			return fmt.Errorf("description was altered since last call to createAnswer")
		}
	default:
		return fmt.Errorf("unknown description type")
	}

	return nil
}

func (pc *RTCPeerConnection) setSessionDescription(description *RTCSessionDescription, remote bool) error {
	if description.SDPType == RTCSdpTypeRollback &&
		(pc.signalingState == RTCSignalingStateStable ||
			pc.signalingState == RTCSignalingStateHaveLocalPranswer ||
			pc.signalingState == RTCSignalingStateHaveRemotePranswer) {
		return makeError(ErrInvalidState, "description type rollback, while signaling state stable/have-local-pranswer/have-remote-pranswer")
	}
	//jsepSetOfTranceivers := pc.rtpTransceivers

	return nil
}

/* func (pc *RTCPeerConnection) setLocalDescription(description *RTCSessionDescription) error {
	var session *sdp.Session
	var sdpType RTCSdpType
	var sdpString string
	var err error

	if description != nil {
		session = description.Session
		sdpType = description.SDPType
		sdpString = description.SDPString
	} else {
		if pc.signalingState == RTCSignalingStateStable || pc.signalingState == RTCSignalingStateHaveLocalOffer || pc.signalingState == RTCSignalingStateHaveRemotePranswer {
			sdpType = RTCSdpTypeOffer
		} else {
			sdpType = RTCSdpTypeAnswer
		}
		session = nil
		sdpString = ""
	}

	if sdpType == RTCSdpTypeOffer && sdpString != "" && sdpString != pc.lastCreatedOffer {
		return makeError(ErrInvalidModification, "description is not equal to last created offer")
	}

	if (sdpType == RTCSdpTypeAnswer || sdpType == RTCSdpTypePranswer) && sdpString != "" && sdpString != pc.lastCreatedAnswer {
		return makeError(ErrInvalidModification, "description is not equal to last created answer")
	}

	if sdpString == "" && sdpType == RTCSdpTypeOffer {
		sdpString = pc.lastCreatedOffer
		if sdpString == "" {
			// TODO: create offer if empty
		} else {
			tmp := sdpString
			buf := bytes.NewBufferString(tmp)

			session, err = sdp.NewDecoder(buf).Decode()
			if err != nil {
				return fmt.Errorf("failed to decode last created offer")
			}
		}
	}

	if sdpString == "" && (sdpType == RTCSdpTypeAnswer || sdpType == RTCSdpTypePranswer) {
		sdpString = pc.lastCreatedAnswer
		if sdpString == "" {
			// TODO(@alisa-vernigor): create answer if empty
		} else {

		}
	}

	return nil
} */

func (pc *RTCPeerConnection) isBundleOnly(bundlePolicy RTCBundlePolicy, isFirstInGroup map[MediaType]bool, kind MediaType) bool {
	if bundlePolicy == RTCBundlePolicyBalancded {
		isFirst := isFirstInGroup[kind]
		isFirstInGroup[kind] = false
		return isFirst
	}

	if bundlePolicy == RTCBundlePolicyMaxCompat {
		return false
	}

	if bundlePolicy == RTCBundlePolicyMaxBundle {
		isFirst := true
		for _, v := range isFirstInGroup {
			isFirst = isFirst && v
		}
		isFirstInGroup[kind] = false
		return isFirst
	}

	return false
}

const (
	alphaNumCharset = "abcdefghijklmnopqrstuvwxyz" + "0123456789"
)

func randStringWithCharset(length int, charset string) string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func (pc *RTCPeerConnection) addAudioCodecs(media *sdp.MediaDesc) {
	media.Fmts = []string{"35"}
	media.AddAttribute("rtpmap", "35 opus/48000")
	media.AddAttribute("fmtp", "35 useinbandfec=1")
}

func (pc *RTCPeerConnection) addVideoCodecs(media *sdp.MediaDesc) {
	media.Fmts = []string{"36"}
	media.AddAttribute("rtpmap", "36 H264 AVC/90000")
	media.AddAttribute("maxptime", "120")
}

func (pc *RTCPeerConnection) addUnmatchedMediaSection(session *sdp.Session, kind MediaType, tlsID string, isBundled bool, fingerprints []*RTCDtlsFingerprint, tranceiver *RTCRtpTransceiver, msidsToMidStrings map[string][]string) {
	if tranceiver.stopped || tranceiver.stopping {
		return
	}

	media := pc.generateMediaSectionBase(isBundled, fingerprints, kind, tlsID)

	pc.addCodecs(media, kind)

	if isBundled {
		media.AddAttribute("bundle-only", " ")
	} else {
		media.AddAttribute("rtcp-rsize", " ")
	}

	for _, mediaStreamId := range tranceiver.sender.associatedMediaStreamIds {
		media.AddAttribute("msid", mediaStreamId)
		msidsToMidStrings[mediaStreamId] = append(msidsToMidStrings[mediaStreamId], strconv.FormatInt(pc.midCounter, 10))
	}

	media.AddAttribute(string(tranceiver.direction), " ")

	session.MediaDescs = append(session.MediaDescs, media)
}

func (pc *RTCPeerConnection) addUnmatchedDataSection(session *sdp.Session,
	mids []string, tlsID string, fingerprints []*RTCDtlsFingerprint) {
	media := pc.generateUnmatchedMediaSectionBase(fingerprints, false, nil, tlsID, mids, MediaTypeApplication)
	media.Port = 9
	media.Fmts = []string{"webrtc-datachannel"}
	media.AddAttribute("sctp-port", "5000")

	session.MediaDescs = append(session.MediaDescs, media)
}

func (pc *RTCPeerConnection) addMediaSections(session *sdp.Session, mids []string, msidsToMidStrings map[string][]string) error {
	isFirstInGroup := map[MediaType]bool{
		MediaTypeAudio:       true,
		MediaTypeVideo:       true,
		MediaTypeText:        true,
		MediaTypeApplication: true,
		MediaTypeMessage:     true,
	}

	tlsID := randStringWithCharset(120, alphaNumCharset)

	var fingerprints []*RTCDtlsFingerprint
	var err error
	for _, cert := range pc.configuration.Certificates {
		fingerprints, err = cert.getFingerprints()
		if err != nil {
			return fmt.Errorf("error while getting fingerprints: %v", err)
		}
	}

	for _, tranceiver := range pc.rtpTransceivers {
		kind := MediaType(tranceiver.sender.track.kind)
		pc.addUnmatchedMediaSection(session, kind, tlsID, pc.isBundleOnly(pc.configuration.BundlePolicy, isFirstInGroup, kind), fingerprints, tranceiver, msidsToMidStrings)
	}

	pc.addUnmatchedDataSection(session, mids, tlsID, fingerprints)

	return nil
}

func (pc *RTCPeerConnection) addMatchedSessionLevelAttributes(session *sdp.Session) {
	pc.addSessionLevelAttributes(session)
	pc.matchIceOptions(session, pc.currentRemoteDescription.Session)
}

func (pc *RTCPeerConnection) addSessionLevelAttributes(session *sdp.Session) {
	session.Version = 0
	session.Originator = &sdp.Origin{
		Username:       "-",
		SessID:         rand.Int63(),
		SessVersion:    pc.sessVersion,
		Nettype:        sdp.NetworkInternet,
		Addrtype:       sdp.TypeIPv4,
		UnicastAddress: "0.0.0.0",
	}
	pc.sessVersion += 1

	session.SessionName = "-"
	session.Timings = []*sdp.Timing{
		{
			Start: 0,
			Stop:  0,
		},
	}

	session.AddAttribute("ice-options", "trickle ice2")
}

func (pc *RTCPeerConnection) createInitialOffer() (*sdp.Session, error) {
	session := &sdp.Session{}
	pc.addSessionLevelAttributes(session)

	mids := []string{}
	msidsToMidStrings := make(map[string][]string)

	err := pc.addMediaSections(session, mids, msidsToMidStrings)
	if err != nil {
		return nil, fmt.Errorf("error while adding mediasections: %v", err)
	}

	if len(mids) > 0 {
		session.AddAttribute("group", "BUNDLE "+strings.Join(mids, " "))
	}

	for _, midStrings := range msidsToMidStrings {
		if len(midStrings) > 1 {
			session.AddAttribute("group", "LS "+strings.Join(midStrings, " "))
		}
	}

	return session, nil
}

func (pc *RTCPeerConnection) CreateOffer(options *RTCOfferOptions) (*RTCSessionDescription, error) {
	if pc.isClosed {
		return nil, makeError(ErrInvalidState, "connection is closed")
	}

	if (pc.signalingState != RTCSignalingStateStable) && (pc.signalingState != RTCSignalingStateHaveLocalOffer) {
		return nil, makeError(ErrInvalidState, "connection's signaling state is neither stable nor have-local-offer")
	}

	var offer *RTCSessionDescription
	for {
		pc.certsGenerated.L.Lock()
		for len(pc.configuration.Certificates) == 0 {
			pc.certsGenerated.Wait()
		}
		pc.certsGenerated.L.Unlock()

		// TODO(@alisa-vernigor): Inspect the offerer's system state to determine the currently available
		// resources as necessary for generating the offer, as described in [RFC8829] (section 4.1.8.)

		offer = &RTCSessionDescription{}
		offer.SDPType = RTCSdpTypeOffer

		tranceivers := pc.rtpTransceivers

		if pc.currentRemoteDescription != nil {
			for _, media := range pc.currentLocalDescription.Session.MediaDescs {
				midArg := media.GetAttribute("mid")
				mid, err := strconv.ParseInt(midArg[0], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("error while getting mid value: %v", err)
				}

				if mid > pc.midCounter {
					pc.midCounter = mid + 1
				}
			}
		}

		if (pc.currentRemoteDescription != nil) || (pc.currentLocalDescription != nil) {
			for _, tranceiver := range tranceivers {
				if tranceiver.stopped {
					continue
				}

				if tranceiver.mid == "" {
					tranceiver.mid = strconv.FormatInt(pc.midCounter, 10)
				}
				pc.midCounter++
			}
		}

		if pc.lastCreatedOffer == "" {
			session, err := pc.createInitialOffer()
			if err != nil {
				return nil, fmt.Errorf("error while creating offer: %v", err)
			} else {
				offer.setSession(session)
				break
			}
		} else {

		}
	}

	return offer, nil
}
