package pulsehelper

import (
	"strings"
	"sync"

	"github.com/function61/gokit/sync/syncutil"
	"github.com/sqp/pulseaudio"
)

// see Work() for why this exists
type Helper struct {
	conn   *pulseaudio.Client
	connMu sync.Mutex
}

// TODO: defer connection making to time of first use?
func New() (*Helper, func() error) {
	pf := &Helper{}
	return pf, pf.Close
}

// wrapper for doing work with PulseAudio client, with reconnection support.
// connection breaks every time PulseAudio is reloaded (think `$ systemctl --user restart pulseaudio`)
func (p *Helper) Work(work func(client *pulseaudio.Client) error) error {
	defer syncutil.LockAndUnlock(&p.connMu)()

	if p.conn == nil {
		// dbus module load check needs to be done every reconnect attempt, because if PulseAudio
		// was restarted (and thus connection break), the module needs to be loaded again.
		isPulseAudioDBusModuleLoaded, err := pulseaudio.ModuleIsLoaded()
		if err != nil {
			return err
		}

		// PulseAudio by default might not have DBus control module loaded, and thus we need to ask it explicitly to load
		if !isPulseAudioDBusModuleLoaded {
			if err := pulseaudio.LoadModule(); err != nil {
				return err
			}
		}

		p.conn, err = pulseaudio.New()
		if err != nil {
			return err
		}
	}

	if err := work(p.conn); err != nil {
		// detect connection breakage
		if strings.Contains(err.Error(), "dbus: connection closed by user") { // FIXME: ugly hack
			p.conn = nil

			// TODO: reconnect here and retry work, but only once?
		}

		return err
	}

	return nil
}

func (p *Helper) Close() error {
	defer syncutil.LockAndUnlock(&p.connMu)()

	if p.conn == nil {
		return nil
	}

	return p.conn.Close()
}
