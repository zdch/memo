package pic

import (
	"github.com/memocash/memo/app/config"
	"os/exec"
	"strconv"
)

func ResizeExternally(from string, to string, width uint, height uint) error {
	var args = []string{
		"--size", strconv.FormatUint(uint64(width), 10) + "x" +
			strconv.FormatUint(uint64(height), 10),
		"--output", to,
		"--crop",
		from,
	}
	path, err := exec.LookPath(config.GetFilePaths().VipsThumbnailPath)
	if err != nil {
		return err
	}
	cmd := exec.Command(path, args...)
	return cmd.Run()
}
