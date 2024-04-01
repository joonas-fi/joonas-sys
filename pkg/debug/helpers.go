package debug

import (
	"context"
	"fmt"

	"github.com/jfreymuth/pulse/proto"
)

func listenSourceChanges(ctx context.Context, sourceName string, processChange func(*proto.GetSourceInfoReply) error) error {
	eventChangesSelectedSource := createSubscribeChangesEventMatcher(proto.EventSource)

	if err := withConn(func(client *proto.Client) error {
		sourceEvents := make(chan interface{}, 1)

		// subscription events are passed via the callback
		client.Callback = func(val interface{}) {
			sourceEvents <- val
		}

		if err := client.Request(&proto.Subscribe{Mask: proto.SubscriptionMaskSource}, nil); err != nil {
			return err
		}

		initialInfo, err := getSourceInfo(proto.Undefined, sourceName, client)
		if err != nil {
			return err
		}

		// trigger initial change notification
		if err := processChange(initialInfo); err != nil {
			return err
		}

		// I guess it's faster to later get info for the source using its index instead of the full name?
		sourceIndex := initialInfo.SourceIndex

		for {
			select {
			case <-ctx.Done():
				return nil
			case eventGeneric := <-sourceEvents:
				if !eventChangesSelectedSource(eventGeneric, sourceIndex) { // we're not interested in this event
					break
				}

				info, err := getSourceInfo(sourceIndex, "", client)
				if err != nil {
					return err
				}

				if err := processChange(info); err != nil {
					return err
				}
			}
		}
	}); err != nil {
		return fmt.Errorf("listenSourceChanges: %w", err)
	}

	return nil
}

func listenSinkChanges(ctx context.Context, sinkName string, processChange func(*proto.GetSinkInfoReply) error) error {
	eventChangesSelectedSink := createSubscribeChangesEventMatcher(proto.EventSink)

	if err := withConn(func(client *proto.Client) error {
		sinkEvents := make(chan interface{}, 1)

		// subscription events are passed via the callback
		client.Callback = func(val interface{}) {
			sinkEvents <- val
		}

		if err := client.Request(&proto.Subscribe{Mask: proto.SubscriptionMaskSink}, nil); err != nil {
			return err
		}

		initialInfo, err := getSinkInfo(proto.Undefined, sinkName, client)
		if err != nil {
			return err
		}

		// trigger initial change notification
		if err := processChange(initialInfo); err != nil {
			return err
		}

		// I guess it's faster to later get info for the source using its index instead of the full name?
		sinkIndex := initialInfo.SinkIndex

		for {
			select {
			case <-ctx.Done():
				return nil
			case eventGeneric := <-sinkEvents:
				if !eventChangesSelectedSink(eventGeneric, sinkIndex) { // we're not interested in this event
					break
				}

				info, err := getSinkInfo(sinkIndex, "", client)
				if err != nil {
					return err
				}

				if err := processChange(info); err != nil {
					return err
				}
			}
		}
	}); err != nil {
		return fmt.Errorf("listenSinkChanges: %w", err)
	}

	return nil
}

// facility is basically either source or sink
func createSubscribeChangesEventMatcher(eventFacility proto.SubscriptionEventType) func(eventGeneric interface{}, index uint32) bool {
	return func(eventGeneric interface{}, index uint32) bool {
		switch event := eventGeneric.(type) {
		case *proto.SubscribeEvent:
			isChangeAndFacilityMatches := event.Event.GetType() == proto.EventChange && event.Event.GetFacility() == eventFacility
			return isChangeAndFacilityMatches && event.Index == index
		default:
			return false
		}
	}
}

func withConn(work func(*proto.Client) error) error {
	// "If the server string is empty, the environment variable PULSE_SERVER will be used."
	client, conn, err := proto.Connect("")
	if err != nil {
		return err
	}
	defer conn.Close()

	/*
		// https://www.freedesktop.org/wiki/Software/PulseAudio/Documentation/Developer/Clients/ApplicationProperties/
			props := proto.PropList{}
			err = client.Request(&proto.SetClientName{Props: props}, nil)
			if err != nil {
				return withErr(err)
			}
	*/

	return work(client)
}

func getSourceInfo(sourceIndex uint32, sourceName string, client *proto.Client) (*proto.GetSourceInfoReply, error) {
	resp := &proto.GetSourceInfoReply{}
	return resp, client.Request(&proto.GetSourceInfo{SourceIndex: sourceIndex, SourceName: sourceName}, resp)
}

func getSinkInfo(sinkIndex uint32, sinkName string, client *proto.Client) (*proto.GetSinkInfoReply, error) {
	resp := &proto.GetSinkInfoReply{}
	return resp, client.Request(&proto.GetSinkInfo{SinkIndex: sinkIndex, SinkName: sinkName}, resp)
}

// returns volume between 0-100
func volumeFor(sinkInfo *proto.GetSinkInfoReply) float64 {
	// TODO: average from all channels
	channelVolume := sinkInfo.ChannelVolumes[0]

	// 0-65536 => 0-100
	return float64(channelVolume) / (float64(proto.VolumeNorm) / 100.0)
}
