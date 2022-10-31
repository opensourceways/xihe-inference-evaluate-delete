package listen

import (
	"errors"
	"github.com/opensourceways/community-robot-lib/utils"
)

type Config struct {
	Inference Inference `json:"inference" required:"true"`
}

type Inference struct {
	RpcEndpoint string `json:"rpc_endpoint" required:"true"`
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
	if c.Inference.RpcEndpoint == "" {
		return errors.New("rpc_endpoint can not be empty")
	}
	return nil
}
