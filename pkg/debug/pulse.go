package debug

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/function61/gokit/app/cli"
	"github.com/jfreymuth/pulse/proto"
	"github.com/spf13/cobra"
)

const (
	defaultSinkID = "@DEFAULT_SINK@"
)

func pulseEntrypoint() *cobra.Command {
	changeVolume := false
	sinkName := defaultSinkID

	cmd := &cobra.Command{
		Use: "pulse-test",
		// Short: "Debug tools",
		Args: cobra.NoArgs,
		Run: cli.RunnerNoArgs(func(ctx context.Context, _ *log.Logger) error {
			switch sinkName {
			case "toggleSourceMute":
				return toggleSourceMute(ctx, "alsa_input.usb-RODE_Microphones_RODE_NT-USB-00.analog-stereo")
			case "listenSourceMute":
				return listenSourceMute(ctx, "alsa_input.usb-RODE_Microphones_RODE_NT-USB-00.analog-stereo")
			case "setDefaultSink":
				return setDefaultSink(ctx, os.Getenv("SINK"))
			case "toggleBetween":
				return toggleBetween(ctx)
			case "listenSinkVolume":
				return listenSinkVolume(ctx, defaultSinkID)
			case "volumeIncrease":
				return volumeIncrease(ctx, defaultSinkID, 5.0)
			case "volumeDecrease":
				return volumeIncrease(ctx, defaultSinkID, -5.0)
			default:
				panic("nein")
			}
		}),
	}

	cmd.Flags().BoolVarP(&changeVolume, "change-volume", "v", changeVolume, "Change volume of the sink")
	cmd.Flags().StringVarP(&sinkName, "sink", "", sinkName, "Sink name")

	return cmd
}

func volumeIncrease(ctx context.Context, sinkName string, by float64) error {
	if err := withConn(func(client *proto.Client) error {
		sinkInfo, err := getSinkInfo(proto.Undefined, sinkName, client)
		if err != nil {
			return err
		}

		currentVolumePct := volumeFor(sinkInfo)

		newVolume := uint32((currentVolumePct + by) / 100.0 * float64(proto.VolumeNorm))

		// TODO: previously we only set slice of one item. is that actually more semantic & officially supported?
		volumes := make([]uint32, len(sinkInfo.ChannelVolumes))
		for i := range volumes {
			volumes[i] = newVolume
		}

		if err := client.Request(&proto.SetSinkVolume{
			SinkIndex:      sinkInfo.SinkIndex,
			SinkName:       "",
			ChannelVolumes: volumes,
		}, nil); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func toggleBetween(ctx context.Context) error {
	speakersDevice := "alsa_output.pci-0000_08_00.1.hdmi-stereo-extra1"
	headphonesDevice := "alsa_output.pci-0000_0a_00.4.analog-stereo"

	return withConn(func(client *proto.Client) error {
		currentDefaultSink, err := getSinkInfo(proto.Undefined, defaultSinkID, client)
		if err != nil {
			return err
		}

		currentlyHeadphones := currentDefaultSink.SinkName == headphonesDevice

		return setDefaultSinkInternal(func() string {
			if currentlyHeadphones {
				return speakersDevice
			} else {
				return headphonesDevice
			}
		}(), client)
	})
}

func setDefaultSink(ctx context.Context, sinkName string) error {
	return withConn(func(client *proto.Client) error {
		return setDefaultSinkInternal(sinkName, client)
	})
}

func setDefaultSinkInternal(sinkName string, client *proto.Client) error {
	if err := client.Request(&proto.SetDefaultSink{
		SinkName: sinkName,
	}, nil); err != nil {
		return fmt.Errorf("setDefaultSinkInternal(sinkName=%s): %w", sinkName, err)
	}

	return nil
}

func toggleSourceMute(ctx context.Context, sourceName string) error {
	if err := withConn(func(client *proto.Client) error {
		current, err := getSourceInfo(proto.Undefined, sourceName, client)
		if err != nil {
			return err
		}

		newMute := !current.Mute // toggle

		if err := client.Request(&proto.SetSourceMute{
			SourceIndex: proto.Undefined,
			SourceName:  sourceName,
			Mute:        newMute,
		}, nil); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return fmt.Errorf("toggleSourceMute: %w", err)
	}

	return nil
}

func listenSourceMute(ctx context.Context, sourceName string) error {
	return listenSourceChanges(ctx, sourceName, func(sourceInfo *proto.GetSourceInfoReply) error {
		fmt.Printf("source mute=%v\n", sourceInfo.Mute)

		return nil
	})
}

func listenSinkVolume(ctx context.Context, sinkName string) error {
	return listenSinkChanges(ctx, sinkName, func(sinkInfo *proto.GetSinkInfoReply) error {
		fmt.Printf("sink volume=%.0f%%\n", volumeFor(sinkInfo))

		return nil
	})
}
