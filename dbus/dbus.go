package dbus

import (
	"log"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/oandrew/ipod"
)

type DeviceDbus interface {
	UpdateTitle(title string)
	UpdateArtist(artist string)
	UpdateAlbum(album string)
	CommitChanges(tr ipod.CommandWriter)
}

func SubscribeDbus() (*dbus.Conn, chan *dbus.Signal) {
	dbus_conn, err := dbus.ConnectSystemBus()
	if err != nil {
		log.Fatalf("could not connect to dbus: %v", err)
	}

	rule := "type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged'"
	err = dbus_conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, rule).Err
	if err != nil {
		log.Fatalf("failed to add dbus match rule: %v", err)
	}

	dbus_channel := make(chan *dbus.Signal, 50)
	dbus_conn.Signal(dbus_channel)

	return dbus_conn, dbus_channel
}

func ProcessDbusUpdate(tr ipod.CommandWriter, dev DeviceDbus, signal *dbus.Signal) {
	if len(signal.Body) < 3 {
		return
	}

	iface, ok := signal.Body[0].(string)
	if !ok || !strings.HasPrefix(iface, "org.bluez") {
		return
	}

	changedProps, ok := signal.Body[1].(map[string]dbus.Variant)
	if !ok {
		return
	}

	if trackInfo, found := changedProps["Track"]; found {
		trackMap, ok := trackInfo.Value().(map[string]dbus.Variant)
		if !ok {
			return
		}

		if titleVariant, ok := trackMap["Title"]; ok {
			if title, ok := titleVariant.Value().(string); ok {
				dev.UpdateTitle(title)
			}
		}
		if artistVariant, ok := trackMap["Artist"]; ok {
			if artist, ok := artistVariant.Value().(string); ok {
				dev.UpdateArtist(artist)
			}
		}
		if albumVariant, ok := trackMap["Album"]; ok {
			if album, ok := albumVariant.Value().(string); ok {
				dev.UpdateAlbum(album)
			}
		}

		dev.CommitChanges(tr)
	}
}
