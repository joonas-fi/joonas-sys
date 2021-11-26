package main

// Generates SSH client configuration file with hostnames filled from Tailscale network

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/function61/gokit/os/osutil"
	"github.com/function61/tailscale-discovery/pkg/tailscalediscoveryclient"
	"github.com/spf13/cobra"
)

func sshConfigGeneratorEntrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "sshconfig-generate",
		Short: "Generate SSH config from Tailscale nodes list",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(sshConfigGenerator(
				osutil.CancelOnInterruptOrTerminate(nil)))
		},
	}
}

func sshConfigGenerator(ctx context.Context) error {
	nodes, err := queryNodesFromTailscale(ctx)
	if err != nil {
		return err
	}

	lines := []string{}

	line := func(input string) { lines = append(lines, input) }

	for _, node := range nodes {
		/*
			Host foobar.example.com
			    HostName foobar.example.com
			    User joonas
		*/
		line(fmt.Sprintf("Host %s", node.hostname))
		line(fmt.Sprintf("    HostName %s", node.ip))
		line(fmt.Sprintf("    User %s", node.sshUsername))
		line("")
	}

	sshConfigFile := strings.Join(lines, "\n")

	// currently need to remove because it's a symlink to Varasto read-only file
	if err := os.Remove("/home/joonas/.ssh/config"); err != nil {
		return err
	}

	if err := os.WriteFile("/home/joonas/.ssh/config", []byte(sshConfigFile), 0600); err != nil {
		return err
	}

	return nil
}

type discoveredNode struct {
	ip          string
	hostname    string
	sshUsername string
}

// TODO: implement local Tailscale API query? there were troubles dialing to the unix socket from host side..
func queryNodesFromTailscale(ctx context.Context) ([]discoveredNode, error) {
	token, err := os.ReadFile("/sto/id/YBZ0O_KRNqE/token")
	if err != nil {
		return nil, err
	}

	discoveryClient := tailscalediscoveryclient.NewClient(
		strings.TrimRight(string(token), "\n"),
		tailscalediscoveryclient.Function61)

	devices, err := discoveryClient.Devices(ctx)
	if err != nil {
		return nil, err
	}

	// curl --unix-socket /var/run/tailscale/tailscaled.sock http://localhost/localapi/v0/status
	// net.Dial()
	// /persist/apps/docker/data_nobackup/overlay2/.../diff/run/tailscale/tailscaled.sock

	discovered := []discoveredNode{}

	for _, device := range devices {
		discovered = append(discovered, discoveredNode{
			ip:          device.IPv4,
			hostname:    device.Hostname,
			sshUsername: guessSSHUsernameFromHostname(device.Hostname),
		})
	}

	// stable iteration order
	sort.Slice(discovered, func(i, j int) bool { return discovered[i].hostname < discovered[j].hostname })

	return discovered, nil
}

// this is really stupid
func guessSSHUsernameFromHostname(hostname string) string {
	switch {
	case strings.HasSuffix(hostname, ".fn61.net"): // Flatcar machines (formerly known as CoreOS)
		return "core"
	case strings.HasSuffix(hostname, "pi"): // henkanpi | veikonpi
		return "pi"
	case hostname == "kodinautomaatio", hostname == "kotomaki":
		return "pi"
	default:
		return "joonas"
	}
}
