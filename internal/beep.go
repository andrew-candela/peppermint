package internal

import (
	"os"
	"os/exec"
)

func Beep() error {
	osa, err := exec.LookPath("osascript")
	if err != nil {
		// Output the only beep we can
		_, err = os.Stdout.Write([]byte{7})
		return err
	}

	cmd := exec.Command(osa, "-e", `beep`)
	return cmd.Run()
}
