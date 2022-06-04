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

func (pc *RTCPeerConnection) isBundleOnly(bundlePolicy RTCBundlePolicy, isFirstInGroup map[string]bool, kind string) bool {
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

func (pc *RTCPeerConnection) generateInitialOffer() (*sdp.Session, error) {
	session := &sdp.Session{}
	session.Version = 0
	session.Originator = &sdp.Origin{
		Username:       "-",
		SessID:         rand.Int63(),
		SessVersion:    pc.sessVersion,
		Nettype:        sdp.NetworkInternet,
		Addrtype:       sdp.TypeIPv4,
		UnicastAddress: "0.0.0.0",
	}
	session.SessionName = "-"
	session.Timings = []*sdp.Timing{
		{
			Start: 0,
			Stop:  0,
		},
	}

	session.AddAttribute("ice-options", "trickle ice2")

	mids := []string{}
	msidsToMidStrings := make(map[string][]string)

	bundlePolicy := pc.configuration.BundlePolicy
	isFirstInGroup := map[string]bool{
		"audio":       true,
		"video":       true,
		"text":        true,
		"application": true,
		"message":     true,
	}

	tlsID := randStringWithCharset(120, alphaNumCharset)
	iceUfrag := randStringWithCharset(4, alphaNumCharset)
	icePwd := randStringWithCharset(22, alphaNumCharset)
	var fingerprints []*RTCDtlsFingerprint
	var err error
	for _, cert := range pc.configuration.Certificates {
		fingerprints, err = cert.getFingerprints()
		if err != nil {
			return nil, fmt.Errorf("error while getting fingerprints: %v", err)
		}
	}

	for _, tranceiver := range pc.rtpTransceivers {
		if tranceiver.stopped || tranceiver.stopping {
			continue
		}
		media := &sdp.MediaDesc{}
		kind := tranceiver.receiver.track.kind

		media.Media = kind

		if tranceiver.sender.track.kind == "audio" {
			media.Fmts = []string{"35"}
			media.AddAttribute("rtpmap", "35 opus/48000")
			media.AddAttribute("fmtp", "35 useinbandfec=1") // Specifies that the decoder has the capability to take advantage of the Opus in-band FEC (Forward Error Correction)
		} else if tranceiver.sender.track.kind == "video" {
			media.Fmts = []string{"36"}
			media.AddAttribute("rtpmap", "36 H264 AVC/90000")
			media.AddAttribute("maxptime", "120")
		}

		if pc.isBundleOnly(bundlePolicy, isFirstInGroup, kind) {
			media.Port = 0
			media.AddAttribute("bundle-only", " ")
		} else {
			media.Port = 9
			media.AddAttribute("rtcp-rsize", " ")
			media.AddAttribute("setup", "actpass")
			media.AddAttribute("tls-id", tlsID)
			media.AddAttribute("ice-ufrag", iceUfrag)
			media.AddAttribute("ice-pwd", icePwd)

			for _, fingerprint := range fingerprints {
				media.AddAttribute("fingerprint", fingerprint.Algorithm+" "+fingerprint.Value)
			}
		}
		media.Proto = []string{"UDP", "TLS", "RTP", "SAVPF"}

		for _, mediaStreamId := range tranceiver.sender.associatedMediaStreamIds {
			media.AddAttribute("msid", mediaStreamId)
			msidsToMidStrings[mediaStreamId] = append(msidsToMidStrings[mediaStreamId], strconv.FormatInt(pc.midCounter, 10))
		}
		media.Connections = []*sdp.Connection{
			{
				Nettype:        sdp.NetworkInternet,
				Addrtype:       sdp.TypeIPv4,
				ConnectionAddr: "0.0.0.0",
				AddressesNum:   1,
			},
		}

		mid := strconv.FormatInt(pc.midCounter, 10)
		media.AddAttribute("mid", mid)
		mids = append(mids, mid)
		pc.midCounter++

		media.AddAttribute(string(tranceiver.direction), " ")
	}

	media := &sdp.MediaDesc{}
	media.Connections = []*sdp.Connection{
		{
			Nettype:        sdp.NetworkInternet,
			Addrtype:       sdp.TypeIPv4,
			ConnectionAddr: "0.0.0.0",
			AddressesNum:   1,
		},
	}
	media.Media = "application"
	media.Proto = []string{"UDP", "DTLS", "SCTP"}
	media.Port = 9
	media.Fmts = []string{"webrtc-datachannel"}
	media.AddAttribute("sctp-port", "5000")

	media.AddAttribute("rtcp-rsize", " ")
	media.AddAttribute("setup", "actpass")
	media.AddAttribute("tls-id", tlsID)
	media.AddAttribute("ice-ufrag", iceUfrag)
	media.AddAttribute("ice-pwd", icePwd)
	for _, fingerprint := range fingerprints {
		media.AddAttribute("fingerprint", fingerprint.Algorithm+" "+fingerprint.Value)
	}

	mid := strconv.FormatInt(pc.midCounter, 10)
	media.AddAttribute("mid", mid)
	mids = append(mids, mid)
	pc.midCounter++

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

	for {
		pc.certsGenerated.L.Lock()
		for len(pc.configuration.Certificates) == 0 {
			pc.certsGenerated.Wait()
		}
		pc.certsGenerated.L.Unlock()

		// TODO(@alisa-vernigor): Inspect the offerer's system state to determine the currently available
		// resources as necessary for generating the offer, as described in [RFC8829] (section 4.1.8.)

	}
}
