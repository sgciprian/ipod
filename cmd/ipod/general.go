package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/davecgh/go-spew/spew"

	"github.com/oandrew/ipod"
	dispremote "github.com/oandrew/ipod/lingo-dispremote"
	extremote "github.com/oandrew/ipod/lingo-extremote"
	general "github.com/oandrew/ipod/lingo-general"

	"github.com/fullsailor/pkcs7"
)

type DevPlayer struct {
	title    string
	artist   string
	album    string
	position uint32
	state    extremote.PlayerState
}

type DevRemoteEventSubscription struct {
	mask uint32
}

type DevGeneral struct {
	uimode    general.UIMode
	tokens    []general.FIDTokenValue
	player    DevPlayer
	remEvents DevRemoteEventSubscription
}

var _ general.DeviceGeneral = &DevGeneral{}

func (d *DevGeneral) UIMode() general.UIMode {
	return d.uimode
}

func (d *DevGeneral) SetUIMode(mode general.UIMode) {
	d.uimode = mode
}

func (d *DevGeneral) Name() string {
	return "ipod-gadget"
}

func (d *DevGeneral) SoftwareVersion() (major uint8, minor uint8, rev uint8) {
	return 7, 1, 2
}

func (d *DevGeneral) SerialNum() string {
	return "abcd1234"
}

func (d *DevGeneral) LingoProtocolVersion(lingo uint8) (major uint8, minor uint8) {
	switch lingo {
	case ipod.LingoGeneralID:
		return 1, 9
	case ipod.LingoDisplayRemoteID:
		return 1, 5
	case ipod.LingoExtRemoteID:
		return 1, 12
	case ipod.LingoDigitalAudioID:
		return 1, 2
	default:
		return 1, 1
	}
}

func (d *DevGeneral) LingoOptions(lingo uint8) uint64 {
	switch lingo {
	case ipod.LingoGeneralID:
		return 0x000000063DEF73FF

	default:
		return 0
	}
}

func (d *DevGeneral) PrefSettingID(classID uint8) uint8 {
	return 0
}

func (d *DevGeneral) SetPrefSettingID(classID uint8, settingID uint8, restoreOnExit bool) {
}

func (d *DevGeneral) SetEventNotificationMask(mask uint64) {

}

func (d *DevGeneral) EventNotificationMask() uint64 {
	return 0
}

func (d *DevGeneral) SupportedEventNotificationMask() uint64 {
	return 0
}

func (d *DevGeneral) CancelCommand(lingo uint8, cmd uint16, transaction uint16) {

}

func (d *DevGeneral) MaxPayload() uint16 {
	return 65535
}

func (d *DevGeneral) StartIDPS() {
	d.tokens = make([]general.FIDTokenValue, 0)
}

func (d *DevGeneral) SetToken(token general.FIDTokenValue) error {
	d.tokens = append(d.tokens, token)
	return nil
}

func (d *DevGeneral) EndIDPS(status general.AccEndIDPSStatus) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Tokens:\n")
	for _, token := range d.tokens {

		fmt.Fprintf(&buf, "* Token: %T\n", token.Token)
		//log.Printf("New token: %T", token.Token)
		switch t := token.Token.(type) {
		case *general.FIDIdentifyToken:

		case *general.FIDAccCapsToken:
			for _, c := range general.AccCaps {
				if t.AccCapsBitmask&uint64(c) != 0 {
					fmt.Fprintf(&buf, "Capability: %v\n", c)
				}
			}
		case *general.FIDAccInfoToken:
			key := general.AccInfoType(t.AccInfoType).String()
			fmt.Fprintf(&buf, "%s: %s\n", key, spew.Sdump(t.Value))

		case *general.FIDiPodPreferenceToken:

		case *general.FIDEAProtocolToken:

		case *general.FIDBundleSeedIDPrefToken:

		case *general.FIDScreenInfoToken:

		case *general.FIDEAProtocolMetadataToken:

		case *general.FIDMicrophoneCapsToken:

		}

	}
	log.Print(buf.String())
}

func (d *DevGeneral) AccAuthCert(cert []byte) {
	pkcs, err := pkcs7.Parse(cert)
	if err != nil {
		log.Error(err)
		return
	}
	if len(pkcs.Certificates) >= 1 {
		cn := pkcs.Certificates[0].Subject.CommonName
		log.Infof("cert: CN=%s", cn)
	}

}

func (d *DevGeneral) PlayingTrackTitle() string {
	return d.player.title
}

func (d *DevGeneral) PlayingTrackArtist() string {
	return d.player.artist
}

func (d *DevGeneral) PlayingTrackAlbum() string {
	return d.player.album
}

var _ extremote.DeviceExtRemote = &DevGeneral{}

func (d *DevGeneral) PlayerState() (uint32, extremote.PlayerState) {
	return d.player.position, d.player.state
}

func (d *DevGeneral) TogglePlayPause(req *ipod.Command, tr ipod.CommandWriter) {
	if d.player.state == extremote.PlayerStatePaused {
		d.player.state = extremote.PlayerStatePlaying
	} else {
		d.player.state = extremote.PlayerStatePaused
	}

	if d.remEvents.mask&(1<<dispremote.RemoteEventPlayStatus) != 0 {
		ipod.Respond(req, tr, &dispremote.RemoteEventNotification{
			EventNum:  dispremote.RemoteEventPlayStatus,
			EventData: []byte{byte(d.PlayStatusType())},
		})
	}
	if d.remEvents.mask&(1<<dispremote.RemoteEventTrackPositionMs) != 0 {
		ipod.Respond(req, tr, &dispremote.RemoteEventNotification{
			EventNum: dispremote.RemoteEventTrackPositionMs,
			EventData: func() []byte {
				b := make([]byte, 4)
				binary.BigEndian.PutUint32(b, d.TrackPositionMs())
				return b
			}(),
		})
	}
}

var _ dispremote.DeviceDispRemote = &DevGeneral{}

func (d *DevGeneral) PlayStatusType() dispremote.PlayStatusType {
	if d.player.state == extremote.PlayerStatePlaying {
		return dispremote.PlayStatusPlaying
	} else {
		return dispremote.PlayStatusPaused
	}
}

func (d *DevGeneral) SetRemoteEventNotificationMask(mask uint32) {
	d.remEvents.mask = mask
}

func (d *DevGeneral) TrackPositionMs() (position uint32) {
	return d.player.position
}

func (d *DevGeneral) TrackInfoTrack() string {
	return d.player.title
}

func (d *DevGeneral) TrackInfoArtist() string {
	return d.player.artist
}

func (d *DevGeneral) TrackInfoAlbum() string {
	return d.player.album
}
