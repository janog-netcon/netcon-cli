package vmms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/janog-netcon/netcon-cli/pkg/types"
	"github.com/sacloud/libsacloud/v2/helper/validate"
	"golang.org/x/xerrors"
)

type client struct {
	Endpoint   string
	Credential string
}

// NewClient vm-management-serverのクライアントを返す
func NewClient(endpoint, credential string) *client {
	return &client{
		Endpoint:   endpoint,
		Credential: credential,
	}
}

type createInstanceRequestBody struct {
	ProblemID        string `json:"problem_id" validate:"required,uuid"`
	MachineImageName string `json:"machine_image_name" validate:"required" example:"problem-sc0"`
}

type createInstanceResponseBody struct {
	Response struct {
		types.Instance
	} `json:"response"`
}

// 今のところ使用していない
type createInstanceErrorResponseBody struct {
	Error struct {
		Code        uint   `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"error"`
}

// CreateInstance VMを作成する
func (c *client) CreateInstance(problemID, machineImageName string) (*types.Instance, error) {
	u := fmt.Sprintf("%s/instance", c.Endpoint)

	reqBody := createInstanceRequestBody{
		ProblemID:        problemID,
		MachineImageName: machineImageName,
	}

	if err := validate.Struct(reqBody); err != nil {
		return nil, err
	}

	reqBodyByte, err := json.Marshal(reqBody)

	req, err := http.NewRequest("GET", u, bytes.NewBuffer(reqBodyByte))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Credential))
	req.Header.Set("Content-Type", "application/json")

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	var body []byte

	// TODO: 200以外の時にbodyに何も入っていなかったらエラーにならないかを確認しておく
	if _, err := resp.Body.Read(body); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, xerrors.New(fmt.Sprintf("status code not 200: %s", body))
	}

	var respBody createInstanceResponseBody
	if err := json.Unmarshal(body, &respBody); err != nil {
		return nil, err
	}

	instance := respBody.Response.Instance

	return &instance, nil
}

type deleteInstanceResponseBody struct {
	Response struct {
		IsDeleted bool `json:"is_deleted"`
	} `json:"response"`
}

// DeleteInstance VMを削除する
func (c *client) DeleteInstance(name string) error {
	u := fmt.Sprintf("%s/instance/%s", c.Endpoint, name)

	req, err := http.NewRequest("GET", u, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Credential))
	req.Header.Set("Content-Type", "application/json")

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}

	var body []byte

	// TODO: 200以外の時にbodyに何も入っていなかったらエラーにならないかを確認しておく
	if _, err := resp.Body.Read(body); err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return xerrors.New(fmt.Sprintf("status code not 200: %s", body))
	}

	var respBody deleteInstanceResponseBody
	if err := json.Unmarshal(body, &respBody); err != nil {
		return err
	}

	if !respBody.Response.IsDeleted {
		return xerrors.New(fmt.Sprintf("delete failed: %s", body))
	}

	return nil
}
