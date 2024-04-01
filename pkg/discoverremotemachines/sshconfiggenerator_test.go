package discoverremotemachines

import (
	"testing"

	"github.com/function61/gokit/testing/assert"
)

func TestPulseAudioReplaceBetween(t *testing.T) {
	const foobar = `#!/usr/bin/pulseaudio -nF

.include /etc/pulse/default.pa

# act as playback server on top of TCP
load-module module-native-protocol-tcp auth-ip-acl=127.0.0.1;192.168.1.0/24;100.64.0.0/10

# <dynamic>
load-module module-tunnel-sink sink_name=work1 server=tcp:100.76.39.10:4713
load-module module-tunnel-sink sink_name=work2 server=tcp:100.76.39.10:4713
# </dynamic>
# list of servers we can send output to (NOTE: might not be safe to include ourselves?)

# yes

`
	assert.EqualString(t, replaceBetween(foobar, "# <dynamic>\n", "# </dynamic>\n", "line 1\nline 2"), `#!/usr/bin/pulseaudio -nF

.include /etc/pulse/default.pa

# act as playback server on top of TCP
load-module module-native-protocol-tcp auth-ip-acl=127.0.0.1;192.168.1.0/24;100.64.0.0/10

# <dynamic>
line 1
line 2# </dynamic>
# list of servers we can send output to (NOTE: might not be safe to include ourselves?)

# yes

`)
}

func TestReplaceBetween(t *testing.T) {
	assert.EqualString(t, replaceBetween(">>>[foobar]<<<", "[", "]", "rust"), ">>>[rust]<<<")
	assert.EqualString(t, replaceBetween("[foobar]", "[", "]", "rust"), "[rust]")
	assert.EqualString(t, replaceBetween("<<<>>>[foobar]<<<", "[", "]", "rust"), "<<<>>>[rust]<<<")
	assert.EqualString(t, replaceBetween("<<<>>>[foobar]<<<", "[", "|", "rust"), "")
	assert.EqualString(t, replaceBetween("<<<>>>[foobar]<<<", "|", "]", "rust"), "")
}
