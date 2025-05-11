package audio

import (
	"io"
	"os"
)

func Persist(stream io.ReadCloser, outPath string) error {
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// No manual buffering, io.Copy automatically copies in 32 KiB chunks
	if _, err := io.Copy(out, stream); err != nil {
		return err
	}
	return out.Sync()
}
