package scoreserver

/*
https://github.com/janog-netcon/netcon-score-server/blob/janog47-changes/vmdb-api/main.go#L85

e.GET("/problem-environments", listProblemEnvironment)
e.GET("/problem-environments/:name", getProblemEnvironment)
*/

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/janog-netcon/netcon-cli/pkg/types"
)

type client struct {
	Endpoint string
}

// NewClient スコアサーバのクライアントを返す
func NewClient(endpoint string) *client {
	return &client{
		Endpoint: endpoint,
	}
}

// ListProblemEnvironment VM一覧を取得する
func (c *client) ListProblemEnvironment() (*[]types.ProblemEnvironment, error) {
	u := fmt.Sprintf("%s/problem-environments", c.Endpoint)

	cli := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	var body []byte

	if _, err := resp.Body.Read(body); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var problemEnvironments []types.ProblemEnvironment
	if err := json.Unmarshal(body, &problemEnvironments); err != nil {
		return nil, err
	}

	return &problemEnvironments, nil
}

// GetProblemEnvironment nameで指定したVMを取得する
func (c *client) GetProblemEnvironment(name string) (*types.ProblemEnvironment, error) {
	u := fmt.Sprintf("%s/problem-environments/%s", c.Endpoint, name)

	cli := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	var body []byte

	if _, err := resp.Body.Read(body); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var problemEnvironment types.ProblemEnvironment
	if err := json.Unmarshal(body, &problemEnvironment); err != nil {
		return nil, err
	}

	return &problemEnvironment, nil
}
