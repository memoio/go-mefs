package role

import (
	"errors"
)

const Debug = false

var (
	ErrNotEnoughBalance = errors.New("balance is insufficient")
	ErrNotMyUser        = errors.New("Not my user")
	ErrNotMyKeeper      = errors.New("Not my keeper")
	ErrNotMyProvider    = errors.New("Not my provider")
	ErrNotKeeper        = errors.New("Role is not a keeper")
	ErrNotProvider      = errors.New("Role is not a provider")
	ErrNotUser          = errors.New("Role is not a user")
	ErrNotConnectd      = errors.New("Role is not connected")

	ErrUkExpire = errors.New("Upkeeping is expired")

	ErrWrongMoney           = errors.New("money is not right")
	ErrWrongSign            = errors.New("signature is not right")
	ErrWrongContarctContent = errors.New("Contract content is wrong")
	ErrWrongKey             = errors.New("Wrong key")
	ErrWrongValue           = errors.New("Wrong value")
	ErrWrongState           = errors.New("Wrong state")

	ErrNoContract      = errors.New("No contract")
	ErrNoBlock         = errors.New("No such block")
	ErrCancel          = errors.New("Cancel")
	ErrRead            = errors.New("Read unexpected err")
	ErrTimeOut         = errors.New("Time out")
	ErrServiceNotReady = errors.New("Service is not ready")

	ErrEmptyData       = errors.New("Data is empty")
	ErrEmptyPrivateKey = errors.New("Empty private key")
	ErrEmptyBlsKey     = errors.New("Empty blskey")

	ErrInvalidInput  = errors.New("Input is invalid")
	ErrInvalidLength = errors.New("Length is invalid")

	ErrTestUser = errors.New("Test user")
)
