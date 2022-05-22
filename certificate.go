package webrtc

type RTCCertificate struct {
	Expires EpochTimeStamp
}

type EpochTimeStamp uint64

func (c *RTCCertificate) getFingerprints() ([]*RTCDtlsFingerprint, error) {
	return nil, nil
}
