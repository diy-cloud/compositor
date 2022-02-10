package register

import (
	"github.com/snowmerak/compositor/vm/multipass"
	"github.com/snowmerak/lux/context"
)

func Get(lc *context.LuxContext) error {
	instance := multipass.New()
	list, err := instance.List()
	if err != nil {
		lc.SetInternalServerError()
		return err
	}

	name := lc.GetPathVariable("name")

	for _, v := range list {
		if v == name {
			lc.SetOK()
			return nil
		}
	}

	lc.SetNotFound()
	return nil
}

func Post(lc *context.LuxContext) error {
	return nil
}
