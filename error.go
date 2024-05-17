package main

import (
	"errors"
	"io"
	"net"
	"strings"
	"syscall"
)

func ConnClosedErr(err error) bool {
	cause := err
	for {
		if unwrap, ok := cause.(interface{ Unwrap() error }); ok {
			cause = unwrap.Unwrap()
			continue
		}
		break
	}

	switch {
	case errors.Is(cause, net.ErrClosed),
		errors.Is(cause, io.EOF),
		errors.Is(cause, syscall.EPIPE):
		return true
	default:
		return strings.Contains(cause.Error(), "forcibly closed by the remote host")

	}
}
