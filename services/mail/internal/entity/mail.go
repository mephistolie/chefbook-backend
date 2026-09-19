package entity

import "errors"

var ErrPermanentDelivery = errors.New("permanent mail delivery failure")

type Mail struct{ To, Subject, Body string }
