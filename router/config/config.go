package config

import (
	"github.com/snowmerak/compositor/vm/multipass"
	"github.com/snowmerak/lux/context"
	"gopkg.in/yaml.v3"
)

func Get(lc *context.LuxContext) error {
	instance := multipass.New()
	info, err := instance.Info(lc.GetPathVariable("id"))
	if err != nil {
		lc.SetBadRequest()
		return err
	}
	data, err := yaml.Marshal(info)
	if err != nil {
		lc.SetInternalServerError()
		return err
	}
	return lc.Reply("application/yaml", data)
}
