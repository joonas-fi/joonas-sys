package statusbar

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/godbus/dbus"
	"github.com/joonas-fi/joonas-sys/pkg/pulsehelper"
	"github.com/sqp/pulseaudio"
)

func getPossibleMicRecordingItem() *barItem {
	if atomic.LoadUint32(&micMuted) == 1 {
		return nil
	}

	// show warning that mic is *NOT* muted

	micRecordingColor := func() string { // flash the text color
		switch time.Now().Unix() % 2 {
		case 1:
			return "#ff0000"
		default:
			return "#c0c0c0"
		}
	}

	return &barItem{
		Name: "rec_indicator",
		// Instance:"",
		// Markup  :"",
		FullText: "🎤 Rec",
		Color:    micRecordingColor(),
	}
}

// actually a bool, but "atomic" package doesn't support bools
var micMuted uint32

func micMonitorTask(ctx context.Context, requestRefresh func()) error {
	pulseHelper, close := pulsehelper.New()
	defer close()

	const target = "alsa_input.usb-RODE_Microphones_RODE_NT-USB-00.analog-stereo"

	mutedStatus := make(chan bool, 1)

	// a bit more than 1 second because our animation color is based on 1 sec, but the ticker time
	// will race and sometimes not hit even seconds so we get "missed" color changes etc.
	const animDuration = 1010 * time.Millisecond
	animTimer := time.NewTicker(animDuration)
	animTimer.Stop()

	return pulseHelper.Work(func(pulse *pulseaudio.Client) error {
		sourceToListen, initialMute, err := micMonitorTaskResolveSource(target, pulse)
		if err != nil { // TODO: handle more gracefully?
			log.Printf("micMonitorTaskResolveSource: %v", err)
			<-ctx.Done()
			return nil
		}

		mutedStatus <- initialMute

		for _, err := range pulse.Register(&pulseSignalsListener{
			sourceToListen: sourceToListen,
			mutedStatus:    mutedStatus,
		}) { // returns multiple errors
			if err != nil {
				return err
			}
		}

		go pulse.Listen()           // dispatches signals to *listener*
		defer pulse.StopListening() // stops above goroutine

		for {
			select {
			case muted := <-mutedStatus:
				atomic.StoreUint32(&micMuted, func() uint32 {
					if muted {
						return 1
					} else {
						return 0
					}
				}())

				if !muted {
					animTimer.Reset(animDuration) // so indicator will keep flashing
				} else {
					animTimer.Stop()
				}

				requestRefresh()
			case <-ctx.Done():
				return nil
			case <-animTimer.C:
				requestRefresh() // to animate indicator
			}
		}
	})
}

type pulseSignalsListener struct {
	sourceToListen dbus.ObjectPath
	mutedStatus    chan bool
}

var _ interface {
	pulseaudio.OnDeviceMuteUpdated
} = (*pulseSignalsListener)(nil)

func (m *pulseSignalsListener) DeviceMuteUpdated(path dbus.ObjectPath, mute bool) {
	if m.sourceToListen != path {
		return
	}

	m.mutedStatus <- mute
}

func micMonitorTaskResolveSource(target string, pulse *pulseaudio.Client) (dbus.ObjectPath, bool, error) {
	sources, err := pulse.Core().ListPath("Sources")
	if err != nil {
		return "", false, err
	}

	for _, source := range sources {
		device := pulse.Device(source)
		deviceName, err := device.String("Name")
		if err != nil {
			return "", false, err
		}

		if deviceName != target {
			continue
		}

		currentMute, err := device.Bool("Mute")
		if err != nil {
			return "", false, err
		}

		return source, currentMute, nil
	}

	return "", false, fmt.Errorf("unable to find source %s", target)
}
