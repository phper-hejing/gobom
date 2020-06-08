package utils

import (
	"github.com/pkg/errors"
	"log"
)

func Err(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(interface{ Cause() error }); ok {
		return err
	}
	return errors.New(err.Error())
}

func ErrPrint(err error) {
	log.Println(errors.Errorf("%+v", err))
}
