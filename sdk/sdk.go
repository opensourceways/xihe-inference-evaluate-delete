package sdk

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"container_manager/controller"
	"github.com/opensourceways/community-robot-lib/utils"
)

type InferenceInfo = controller.InferenceInfo

type InferenceEvaluate struct {
	endpoint string
	cli      utils.HttpClient
}

func NewInferenceEvaluate(endpoint string) InferenceEvaluate {
	return InferenceEvaluate{
		endpoint: strings.TrimSuffix(endpoint, "/"),
		cli:      utils.NewHttpClient(3),
	}
}

func (i InferenceEvaluate) InferenceCreate(info *InferenceInfo) error {
	payload, err := utils.JsonMarshal(info)
	if err != nil {
		return err
	}

	url := i.endpoint + "/inference/create"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	if err = i.forwardTo(req, nil); err != nil {
		return err
	}
	return nil
}

func (i InferenceEvaluate) InferenceExpiry(info *InferenceInfo, expiry int) error {
	info.Expiry = strconv.Itoa(expiry)
	payload, err := utils.JsonMarshal(info)
	if err != nil {
		return err
	}

	url := i.endpoint + "/inference/extend_expiry"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	if err = i.forwardTo(req, nil); err != nil {
		return err
	}
	return nil
}

func (i InferenceEvaluate) forwardTo(req *http.Request, jsonResp interface{}) (err error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "xihe-inference-evaluate")

	if jsonResp != nil {
		v := struct {
			Data interface{} `json:"data"`
		}{jsonResp}

		_, err = i.cli.ForwardTo(req, &v)
	} else {
		_, err = i.cli.ForwardTo(req, jsonResp)
	}
	return
}
