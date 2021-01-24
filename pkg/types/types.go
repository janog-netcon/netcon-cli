package types

import (
	"time"

	"github.com/gofrs/uuid"
)

/*
https://github.com/janog-netcon/netcon-score-server/blob/bfff5171e739801e0c129c11cc681276b66c8af8/api/db/schema.rb#L131
```ruby
create_table "problem_environments", id: :uuid, default: -> { "gen_random_uuid()"  }, force: :cascade do |t|
	t.string "status"
	t.string "host", null: false
	t.string "user", null: false
	t.string "password", null: false
	t.uuid "problem_id", null: false
	t.uuid "team_id"
	t.datetime "created_at", null: false
	t.datetime "updated_at", null: false
	t.string "secret_text", null: false
	t.string "name", null: false
	t.string "service", null: false
	t.integer "port", null: false
	t.string "machine_image_name"
	t.string "external_status"
	t.index ["problem_id", "name", "service"], name: "problem_environments_on_composit_keys", unique: true
	t.index ["problem_id"], name: "index_problem_environments_on_problem_id"
	t.index ["team_id"], name: "index_problem_environments_on_team_id"
end
```

```
class ProblemEnvironment < ApplicationRecord
    # インターフェースから見るとname, service, team, problemが複合プライマリーキー
	validates :problem,  presence: true, uniqueness: { scope: %i[team_id name service]  }
    # teamがnilなら共通
	validates :team,     presence: false
	validates :name,     presence: true
	validates :service,  presence: true

	validates :status,   presence: false
	validates :host,     presence: false
	validates :port,     presence: true, numericality: { only_integer: true  }
	validates :user,     presence: false
	validates :password,    allow_empty: true
	validates :secret_text, allow_empty: true

	belongs_to :team, optional: true
	belongs_to :problem
end
```

レコード例:
```
id                  | status |     host      |   user   | password |              problem_id              | team_id |        created_at         |         updated_at         |    secret_text    |      name       | service | port  | machine_image_name | external_status
--------------------------------------+--------+---------------+----------+----------+--------------------------------------+---------+---------------------------+----------------------------+-------------------+-----------------+---------+-------+--------------------+-----------------
21864669-7eed-42df-98e3-e96e2c5857b0 |        | 35.187.220.33 | j47-user | SjUVfdKB | 4b71d7be-6a76-4a10-a16b-9f50b47c3407 |         | 2021-01-07 21:43:07.13899 | 2021-01-07 22:06:06.069066 | generated by vmms | image-110-okaxv | SSH     | 50080 | image-110          | STOPPING
71ff4819-fd3d-4ca6-8598-37c2c365c70f |        | 35.187.220.33 | j47-user | SjUVfdKB | 4b71d7be-6a76-4a10-a16b-9f50b47c3407 |         | 2021-01-07 21:43:07.18566 | 2021-01-07 22:06:06.09428  | generated by vmms | image-110-okaxv | HTTPS   |   443 | image-110          | STOPPING
```
*/

const (
	// ProblemEnvironmentInnerStatusNotReady 準備中
	ProblemEnvironmentInnerStatusNotReady = "NOT_READY"
	// ProblemEnvironmentInnerStatusReady プール中
	ProblemEnvironmentInnerStatusReady = "READY"
	// ProblemEnvironmentInnerStatusUnderChallenge ユーザが解答中
	ProblemEnvironmentInnerStatusUnderChallenge = "UNDER_CHALLENGE"
	// ProblemEnvironmentInnerStatusUnderScoring 採点中
	ProblemEnvironmentInnerStatusUnderScoring = "UNDER_SCORING"
	// ProblemEnvironmentInnerStatusAbandoned 削除中
	ProblemEnvironmentInnerStatusAbandoned = "ABANDONED"
)

// ProblemEnvironment スコアサーバが管理しているVM情報
type ProblemEnvironment struct {
	ID               uuid.UUID `json:"id"`
	InnerStatus      *string   `json:"inner_status"`
	Status           *string   `json:"status"`
	Host             string    `json:"host"`
	User             string    `json:"user"`
	Password         string    `json:"password"`
	ProblemID        string    `json:"problem_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	ProjectName      string    `json:"project"`
	ZoneName         string    `json:"zone"`
	Name             string    `json:"name"`
	Service          string    `json:"service"`
	Port             int       `json:"port"`
	MachineImageName *string   `json:"machine_image_name"`
}

// Instance vm-management-serverから返ってくるinstanceのobject
type Instance struct {
	InstanceName     string `json:"instance_name" validate:"required" example:"problem-sc0-li5qj"`
	MachineImageName string `json:"machine_image_name" validate:"required" example:"problem-sc0"`
	Domain           string `json:"domain" validate:"required" example:"xxxxxxxx.janog47.eve-ng.com"`
	Status           string `json:"status" validate:"required" example:"RUNNING"`
	ProblemID        string `json:"problem_id" validate:"required,uuid" example:"uuid"`
	UserID           string `json:"user_id" validate:"required" example:"j47-user"`
	Password         string `json:"password" validate:"required" example:"xxxxxxxx"`
}

// SchedulerConfig schedulerの設定ファイルで使用する
type SchedulerConfig struct {
	Setting struct {
		Scoreserver struct {
			Endpoint string `yaml:"endpoint"`
		} `yaml:"scoreserver"`
		Vmms struct {
			Endpoint   string `yaml:"endpoint"`
			Credential string `yaml:"credential"`
		} `yaml:"vmms"`
		Cron     string `yaml:"cron"`
		Projects []struct {
			Name  string `yaml:"name"`
			Zones []struct {
				Name        string `yaml:"name"`
				MaxInstance int    `yaml:"max_instance"`
				Priority    int    `yaml:"priority"`
			} `yaml:"zones"`
		} `yaml:"projects"`
		Problems []struct {
			Name            string `yaml:"name"`
			KeepPool        int    `yaml:"keep_pool"`
			DefaultInstance int    `yaml:"default_instance"`
		} `yaml:"problems"`
	} `yaml:"setting"`
}

//問題ごとの情報
type ProblemInstance struct {
	MachineImageName string
	ProblemID        string
	NotReady         int
	Ready            int
	UnderChallenge   int
	UnderScoring     int
	Abandoned        int
	KeepPool         int
	KIS              []KeepInstance
	CurrentInstance  int
	DefaultInstance  int
}

type KeepInstance struct {
	InstanceName string
	ProjectName  string
	ZoneName     string
	CreatedAt    time.Time
}

type ZonePriority struct {
	ProjectName     string
	ZoneName        string
	Priority        int
	MaxInstance     int
	CurrentInstance int
}

type CreateInstance struct {
	ProblemName      string
	ProblemID        string
	MachineImageName string
}

type DeleteInstance struct {
	ProblemName  string
	InstanceName string
	ProjectName  string
	ZoneName     string
}
