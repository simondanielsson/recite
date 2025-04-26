package player

import (
	"io"
	"time"

	"github.com/ebitengine/oto/v3"
)

// Play plays a stream out loud.
func Play(stream io.ReadCloser) error {
	op := &oto.NewContextOptions{
		SampleRate:   24000,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
	}

	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return err
	}
	<-readyChan

	player := otoCtx.NewPlayer(stream)
	player.Play()
	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}
	err = player.Close()
	if err != nil {
		return err
	}
	return nil
}
