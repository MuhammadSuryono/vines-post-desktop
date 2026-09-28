//go:build !darwin

package main

import (
	"io"
	"os"
)

func openPrinter(printerName string) (io.WriteCloser, error) {
	name, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	path := "\\\\" + name + "\\" + printerName
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)
}
