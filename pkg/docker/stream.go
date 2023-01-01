package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/cli/cli"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/progress"
	log "github.com/sirupsen/logrus"
)

type message jsonmessage.JSONMessage

func (m message) String() string {
	m.Stream = strings.TrimSpace(m.Stream)
	bf := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(bf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(m)
	return strings.TrimSpace(bf.String())
}

func decodeStream(reader io.ReadCloser) error {
	dec := json.NewDecoder(reader)
	for {
		var jm message
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if msg := jm.String(); msg != "{}" {
			log.Debug(msg)
		}
		if jm.Error != nil {
			return cli.StatusError{Status: jm.Error.Message, StatusCode: jm.Error.Code}
		}
	}
	return nil
}

type formatProgress interface {
	formatStatus(id, format string, a ...interface{}) []byte
	formatProgress(id, action string, progress *jsonmessage.JSONProgress) []byte
}

type rawProgressFormatter struct{}

func (sf *rawProgressFormatter) formatStatus(id, format string, a ...interface{}) []byte {
	return []byte(fmt.Sprintf(format, a...))
}

func (sf *rawProgressFormatter) formatProgress(id, action string, progress *jsonmessage.JSONProgress) []byte {
	return []byte(action + " " + progress.String())
}

type progressLog struct {
	sf formatProgress
}

// WriteProgress formats progress information from a ProgressReader.
func (out *progressLog) WriteProgress(prog progress.Progress) error {
	var formatted []byte
	if prog.Message != "" {
		formatted = out.sf.formatStatus(prog.ID, prog.Message)
	} else {
		jsonProgress := jsonmessage.JSONProgress{
			Current:    prog.Current,
			Total:      prog.Total,
			HideCounts: prog.HideCounts,
			Units:      prog.Units,
		}
		formatted = out.sf.formatProgress(prog.ID, prog.Action, &jsonProgress)
	}
	log.Debug(string(formatted))
	return nil
}
