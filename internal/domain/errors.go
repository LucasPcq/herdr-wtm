package domain

import "errors"

var (
	ErrWtmTooOld       = errors.New("herdr-wtm needs wtm 0.29 or later: run `wtm upgrade`")
	ErrWtmSchemaTooNew = errors.New("wtm events stopped on an event newer than this wtm understands: run `wtm upgrade`")
)
