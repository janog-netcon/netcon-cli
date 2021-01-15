package scoreserver

/*
https://github.com/janog-netcon/netcon-score-server/blob/janog47-changes/vmdb-api/main.go#L85

e.GET("/problem-environments", listProblemEnvironment)
e.GET("/problem-environments/:name", getProblemEnvironment)
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/janog-netcon/netcon-cli/pkg/types"
	"golang.org/x/xerrors"
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
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, xerrors.New("status code not 200")
	}

	var problemEnvironments []types.ProblemEnvironment
	if err := json.Unmarshal(respBody, &problemEnvironments); err != nil {
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
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, xerrors.New("status code not 200")
	}

	var problemEnvironment types.ProblemEnvironment
	if err := json.Unmarshal(respBody, &problemEnvironment); err != nil {
		return nil, err
	}

	return &problemEnvironment, nil
}
