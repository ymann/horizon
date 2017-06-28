package scenarios

import (
	"bytes"
	"os/exec"
	"github.com/openbankit/horizon/log"
	"bufio"
	"os"
)

//go:generate go-bindata -ignore (go|rb)$ -pkg scenarios .

// Load executes the sql script at `path` on postgres database at `url`
func Load(url string, path string) {
	sql, err := Asset(path)

	if err != nil {
		log.WithField("service", "load_scenarious").WithError(err).Panic("Failed to load scenario")
	}

	psql := exec.Command("psql", url)
	psql.Stdin = bytes.NewReader(sql)
	psqlErr := runCommand(psql)

	if psqlErr != nil {
		log.Panic("Failed to load scenario")
	}

}

func runCommand(cmd *exec.Cmd) error {
	w := bufio.NewWriter(os.Stdout)
	cmd.Stdout = w
	cmd.Stderr = w
	err := cmd.Run()
	if err != nil {
		log := log.WithField("service", "load_scenarious")
		flushErr := w.Flush()
		if flushErr != nil {
			log.WithError(flushErr).Error("Failed to flush error data")
		}
	}
	return err
}

