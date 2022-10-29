package listen

import (
	"errors"
	"github.com/opensourceways/community-robot-lib/utils"
)

type Config struct {
	Inference Inference `json:"inference" required:"true"`
}

type Inference struct {
	NotifyUrl string `json:"notify_url" required:"true"`
}

func loadConfig(nConfig *Config) error {
	if err := utils.LoadFromYaml("./config.yaml", nConfig); err != nil {
		return err
	}
	if err := nConfig.Validate(); err != nil {
		return err
	}
	return nil
}

func (c Config) Validate() error {
	if c.Inference.NotifyUrl == "" {
		return errors.New("notify_url can not be empty")
	}
	return nil
}
