package commands

import (
	"fmt"
	"io"

	"github.com/cheggaaa/pb"
)

type clearlineReader struct {
	io.Reader
	out io.Writer
}

func (r *clearlineReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	if err == io.EOF {
		// callback
		fmt.Fprintf(r.out, "\033[2K\r") // clear progress bar line on EOF
	}
	return
}

func progressBarForReader(out io.Writer, r io.Reader, l int64) (*pb.ProgressBar, io.Reader) {
	bar := makeProgressBar(out, l)
	barR := bar.NewProxyReader(r)
	return bar, &clearlineReader{barR, out}
}

func makeProgressBar(out io.Writer, l int64) *pb.ProgressBar {
	// setup bar reader
	// TODO: get total length of files
	bar := pb.New64(l).SetUnits(pb.U_BYTES)
	bar.Output = out

	// the progress bar lib doesn't give us a way to get the width of the output,
	// so as a hack we just use a callback to measure the output, then get rid of it
	bar.Callback = func(line string) {
		terminalWidth := len(line)
		bar.Callback = nil
		log.Infof("terminal width: %v\n", terminalWidth)
	}
	return bar
}
