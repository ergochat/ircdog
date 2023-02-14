package lib

import (
	"fmt"
	"os"
	"sync"
)

type Transcript struct {
	sync.Mutex
	outfile *os.File
}

func NewTranscript(filename string) (result *Transcript, err error) {
	outfile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	return &Transcript{
		outfile: outfile,
	}, nil
}

func (t *Transcript) Close() error {
	if t == nil {
		return nil
	}
	return t.outfile.Close()
}

func (t *Transcript) WriteLine(line string, isClient bool) (err error) {
	if t == nil {
		return nil
	}
	marker := "<- "
	if isClient {
		marker = "-> "
	}
	t.Lock()
	defer t.Unlock()
	// XXX due to an implementation limitation of ircreader, we are effectively normalizing
	// terminating \n to \r\n even when the \r was absent:
	_, err = fmt.Fprintf(t.outfile, "%s%s\r\n", marker, line)
	return
}
