package client

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/darmiel/yaxc/internal/api"
	"github.com/darmiel/yaxc/internal/common"
	"github.com/muesli/termenv"
	"time"
)

func WatchServer(c *Check, d time.Duration, done chan bool) {
	t := time.NewTicker(d)
	for {
		select {
		case <-done:
			fmt.Println(common.StyleInfo(), "Stopping", common.WordServer(), "Watcher")
			return
		case <-t.C:
			if err := c.CheckServer(); err != nil {
				fmt.Println(common.StyleWarn(), err)
			}
			break
		}
	}
}

func (c *Check) CheckServer() (err error) {
	// check if clipboard actions are available
	if clipboard.Unsupported {
		return ErrUnsupported
	}

	// lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// get hash from server
	var sh string
	if sh, err = c.a.GetHash(c.path); err != nil {
		if err == api.ErrErrResponse {
			err = nil
		}
		return
	}
	if sh == "" {
		return ErrEmptyHash
	}

	// get clipboard
	var cb string
	cb, _ = common.GetClipboard(c.useBase64)
	empty := cb == ""

	if c.useBase64 {
		cb = base64.StdEncoding.EncodeToString([]byte(cb))
	}

	if !empty {
		// calculate hash
		var ch string
		if ch = common.MD5Hash(cb); ch == "" {
			err = ErrEmptyHash
			return
		}

		// compare hashes
		if ch == sh {
			return
		}
	}

	// compare last update
	if c.previousClipboard != cb {
		return errors.New("ignored because previous clipboard differ")
	}

	// get data from server
	var sd string
	if sd, err = c.a.GetContent(c.path, c.pass); err != nil {
		return
	}

	if c.useBase64 {
		var b []byte
		if b, err = base64.StdEncoding.DecodeString(sd); err != nil {
			return
		}
		sd = string(b)
		fmt.Println(common.StyleInfo(), "Decoded clipboard to",
			termenv.String(sd).Foreground(common.Profile().Color("#66C2CD")))
	}

	// update contents
	c.previousClipboard = sd
	err = common.WriteClipboard(sd, c.useBase64)
	fmt.Println(common.StyleUpdate(), common.WordClient(), "<-", common.PrettyLimit(cb, 48))
	return
}
