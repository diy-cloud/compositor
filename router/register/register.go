package register

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/snowmerak/compositor/compress"
	"github.com/snowmerak/compositor/config"
	"github.com/snowmerak/lux/context"
)

func Post(lc *context.LuxContext) error {
	body, err := lc.GetBody()
	if err != nil {
		lc.SetBadRequest()
		return err
	}

	homePath := filepath.Join(config.HomePath, lc.GetPathVariable("id"))
	if err := os.RemoveAll(homePath); err != nil {
		if !os.IsNotExist(err) {
			lc.SetInternalServerError()
			return err
		}
	}
	if err := os.MkdirAll(homePath, 0755); err != nil {
		lc.SetInternalServerError()
		return err
	}
	if err := compress.Untar(bytes.NewReader(body), homePath); err != nil {
		lc.SetInternalServerError()
		return err
	}

	buf := [64]byte{}
	if _, err := rand.Read(buf[:]); err != nil {
		lc.SetInternalServerError()
		return err
	}

	name := lc.GetPathVariable("id") + "-" + hex.EncodeToString(buf[:])

	return nil
}
