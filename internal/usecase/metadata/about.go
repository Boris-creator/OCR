package metadata

import (
	"fmt"
	"os/exec"
)

type About struct{}

func New() *About {
	return &About{}
}

func (service About) GetSourceCodeURL() (string, error) {
	const errPrefix = "GetSourceCodeURL"

	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	out, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("%s: command.Run %s: %s; %w", errPrefix, cmd.String(), out, err)
	}

	return string(out), nil
}
