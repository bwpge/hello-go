package server

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/matoous/go-nanoid/v2"
)

const validDuration = time.Second * 5

type Otp struct {
	value   string
	created time.Time
}

func NewOtp() *Otp {
	value, err := gonanoid.New()
	if err != nil {
		log.Fatalf("error generating OTP: %v", err)
	}

	return &Otp{
		value:   value,
		created: time.Now(),
	}
}

func (o *Otp) IsExpired() bool {
	return time.Now().Sub(o.created) > validDuration
}

func (o *Otp) Validate(value string) bool {
	if o.IsExpired() {
		return false
	}

	return o.value == value
}
