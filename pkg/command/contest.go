package command

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/janog-netcon/netcon-cli/pkg/vmms"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v2"
)

func NewContestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "contest",
	}

	cmd.AddCommand(
		NewContestInitCommand(),
	)

	flags := cmd.PersistentFlags()
	flags.StringP("vmms-endpoint", "", "http://127.0.0.1:8950", "vm-management-server Endpoint")
	flags.StringP("vmms-credential", "", "", "Token")

	return cmd
}

func NewContestInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "init",
		RunE: contestInitCommandFunc,
	}

	flags := cmd.Flags()
	flags.StringP("mapping-file-path", "", "", "problem-idとmachine-image-idのマッピング情報が書いてあるファイルを指定する")
	flags.UintP("count", "c", 1, "何台ずつ作成するか")

	cmd.MarkFlagRequired("mapping-file-path")

	return cmd
}

type mapping struct {
	ProblemID        string `yaml:"problem_id"`
	MachineImageName string `yaml:"machine_image_name"`
	Project          string `yaml:"project"`
	Zone             string `yaml:"zone"`
}

func contestInitCommandFunc(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	vmmsEndpoint, err := flags.GetString("vmms-endpoint")
	if err != nil {
		return err
	}
	vmmsCredential, err := flags.GetString("vmms-credential")
	if err != nil {
		return err
	}
	mappingFilePath, err := flags.GetString("mapping-file-path")
	if err != nil {
		return err
	}
	count, err := flags.GetUint("count")
	if err != nil {
		return err
	}

	// read mapping file
	bytes, err := ioutil.ReadFile(mappingFilePath)
	if err != nil {
		return err
	}

	ml := []mapping{}
	if err := yaml.Unmarshal(bytes, &ml); err != nil {
		return err
	}

	fmt.Printf("[INFO] read success: %#v\n", ml)

	// validate
	for _, m := range ml {
		if m.ProblemID == "" {
			xerrors.New("problem_id が空になっている場所があります")
		}
		if m.MachineImageName == "" {
			xerrors.New("machine-image-name が空になっている場所があります")
		}
		if m.Project == "" {
			xerrors.New("project が空になっている場所があります")
		}
		if m.Zone == "" {
			xerrors.New("zone が空になっている場所があります")
		}
	}

	// create instance
	cli := vmms.NewClient(vmmsEndpoint, vmmsCredential)

	// 問題ごとに指定カウント分作成させた方が、途中でコケたときに扱いやすい
	// 作成が完了した問題は設定ファイルから削除すればよくなる
	for _, m := range ml {
		c := count

		for c > 0 {
			// TODO: リトライカウントの実装
			// 今は作成に成功するまで無限ループ
			for {
				fmt.Printf("[INFO] creating... problemID: %s, machineImageName: %s, project: %s, zone: %s\n", m.ProblemID, m.MachineImageName, m.Project, m.Zone)

				i, err := cli.CreateInstance(m.ProblemID, m.MachineImageName, m.Project, m.Zone)
				if err != nil {
					fmt.Println("[ERROR] failed to create instance.")
					// VMの作成に失敗した場合は5秒sleepする
					time.Sleep(time.Second * 5)
				} else {
					fmt.Printf("[INFO] created: %#v\n", i)
					break
				}
			}

			c--
		}
	}

	fmt.Println("[INFO] success!!!!")

	return nil
}
