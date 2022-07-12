package vmms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/janog-netcon/netcon-cli/pkg/types"
	"github.com/sacloud/libsacloud/v2/helper/validate"
	"golang.org/x/xerrors"
)

type Client struct {
	Endpoint   string
	Credential string
}

// NewClient vm-management-serverのクライアントを返す
func NewClient(endpoint, credential string) *Client {
	return &Client{
		Endpoint:   endpoint,
		Credential: credential,
	}
}

type createInstanceRequestBody struct {
	ProblemID        string `json:"problem_id" validate:"required,uuid"`
	MachineImageName string `json:"machine_image_name" validate:"required" example:"problem-sc0"`
	Project          string `json:"project" validate:"required" example:"networkcontest"`
	Zone             string `json:"zone" validate:"required" zone:"asia-northeast1-b"`
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
func (c *Client) CreateInstance(problemID, machineImageName, project, zone string) (*types.Instance, error) {
	u := fmt.Sprintf("%s/instance", c.Endpoint)

	reqBody := createInstanceRequestBody{
		ProblemID:        problemID,
		MachineImageName: machineImageName,
		Project:          project,
		Zone:             zone,
	}

	if err := validate.Struct(reqBody); err != nil {
		return nil, err
	}

	reqBodyByte, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u, bytes.NewBuffer(reqBodyByte))
	if err != nil {
		return nil, err
	}
	// https://stackoverflow.com/questions/17714494/golang-http-request-results-in-eof-errors-when-making-multiple-requests-successi
	// これが無いとエラーになる `Failed to CreateInstance. Post "http://vm-management-service:81/instance": EOF``
	req.Close = true
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Credential))
	req.Header.Set("Content-Type", "application/json")

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// TODO: 200以外の時にbodyに何も入っていなかったらエラーにならないかを確認しておく
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, xerrors.New(fmt.Sprintf("status code not 200: status code is %d: body: %s", resp.StatusCode, body))
	}

	var respBody createInstanceResponseBody
	if err := json.Unmarshal(body, &respBody); err != nil {
		return nil, xerrors.Errorf("body %s:json unmarshal error: %w", respBody, err)
	}

	instance := respBody.Response.Instance

	return &instance, nil
}

type deleteInstanceRequestBody struct {
	Project string `json:"project" validate:"required" example:"networkcontest"`
	Zone    string `json:"zone" validate:"required" zone:"asia-northeast1-b"`
}

type deleteInstanceResponseBody struct {
	Response struct {
		IsDeleted bool `json:"is_deleted"`
	} `json:"response"`
}

// DeleteInstance VMを削除する
func (c *Client) DeleteInstance(name, project, zone string) error {
	u := fmt.Sprintf("%s/instance/%s", c.Endpoint, name)

	reqBody := deleteInstanceRequestBody{
		Project: project,
		Zone:    zone,
	}

	if err := validate.Struct(reqBody); err != nil {
		return err
	}

	reqBodyByte, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", u, bytes.NewBuffer(reqBodyByte))
	if err != nil {
		return err
	}
	// https://stackoverflow.com/questions/17714494/golang-http-request-results-in-eof-errors-when-making-multiple-requests-successi
	// これが無いとエラーになる `Failed to CreateInstance. Post "http://vm-management-service:81/instance": EOF``
	req.Close = true
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Credential))
	req.Header.Set("Content-Type", "application/json")

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return xerrors.New(fmt.Sprintf("status code not 200: status code is %d: body: %s", resp.StatusCode, body))
	}

	return nil
}
