package mpresque

import (
	"flag"
	"fmt"
	"log"

	"github.com/go-redis/redis"
	mp "github.com/mackerelio/go-mackerel-plugin-helper"
)

var graphdef = map[string]mp.Graphs{
	"queues": {
		Label: "Resque queues",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "pending_sum", Label: "Sum pending count", Diff: false, Stacked: true},
		},
	},
	"queue.#": {
		Label: "Resque queue",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "pending", Label: "Pending", Diff: false, Stacked: true},
		},
	},
	"worker": {
		Label: "Resque worker",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "processes", Label: "Processes", Diff: false},
		},
	},
	"stat": {
		Label: "Resque stat",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "processed", Label: "Job processed count", Diff: true},
			{Name: "failed", Label: "Job failed count", Diff: true},
		},
	},
}

// ResquePlugin mackerel plugin for Resque
type ResquePlugin struct {
	Prefix    string
	Namespace string
	Host      string
	Port      string
	Password  string
	DB        int
	Redis     *redis.Client
	Queues    []string
}

func (r *ResquePlugin) prepare() error {
	addr := fmt.Sprintf("%s:%s", r.Host, r.Port)
	r.Redis = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: r.Password,
		DB:       r.DB,
		Network:  "tcp",
	})

	_, err := r.Redis.Ping().Result()
	if err != nil {
		return err
	}

	queuesKey := fmt.Sprintf("%s:%s", r.Namespace, "queues")
	r.Queues, err = r.Redis.SMembers(queuesKey).Result()
	if err != nil {
		return err
	}

	return nil
}

// FetchMetrics interface for mackerelplugin
func (r ResquePlugin) FetchMetrics() (map[string]interface{}, error) {
	ret := make(map[string]interface{})

	var pendingSum int64
	pendingSum = 0

	for _, q := range r.Queues {

		qKey := fmt.Sprintf("%s:%s:%s", r.Namespace, "queue", q)
		qlen, err := r.Redis.LLen(qKey).Result()
		if err != nil {
			return nil, err
		}
		ret["queue."+q+".pending"] = float64(qlen)
		pendingSum += qlen

	}
	ret["pending_sum"] = float64(pendingSum)

	workerKey := fmt.Sprintf("%s:%s", r.Namespace, "workers")
	workerProcesses, err := r.Redis.SCard(workerKey).Result()
	if err != nil {
		return nil, err
	}
	ret["processes"] = float64(workerProcesses)

	failedKey := fmt.Sprintf("%s:%s", r.Namespace, "stat:failed")
	failedCount, err := r.Redis.Get(failedKey).Float64()
	if err != nil {
		return nil, err
	}
	ret["failed"] = failedCount

	processedKey := fmt.Sprintf("%s:%s", r.Namespace, "stat:processed")
	processedCount, err := r.Redis.Get(processedKey).Float64()
	if err != nil {
		return nil, err
	}
	ret["processed"] = processedCount

	return ret, nil

}

// GraphDefinition interface for mackerelplugin
func (r ResquePlugin) GraphDefinition() map[string]mp.Graphs {
	graphdef := graphdef

	var queuesMetrics []mp.Metrics
	for _, v := range r.Queues {
		m := mp.Metrics{
			Name:    v,
			Label:   v,
			Diff:    false,
			Stacked: true,
		}
		queuesMetrics = append(queuesMetrics, m)
	}

	graphdef["pending"] = mp.Graphs{
		Label:   "Resque queue",
		Unit:    "integer",
		Metrics: queuesMetrics,
	}

	return graphdef
}

// MetricKeyPrefix interface for PluginWithPrefix
func (r ResquePlugin) MetricKeyPrefix() string {
	if r.Prefix == "" {
		r.Prefix = "resque"
	}
	return r.Prefix
}

// Do the plugin
func Do() {
	var (
		optPrefix    = flag.String("metric-key-prefix", "resque", "Metric key prefix")
		optNamespace = flag.String("namespace", "resque", "Redis key prefix")
		optHost      = flag.String("host", "127.0.0.1", "The bind url to use for the redis server")
		optPort      = flag.String("port", "6379", "The bind port to use for the redis server")
		optPassword  = flag.String("password", "", "Password for the redis server")
		optDB        = flag.Int("db", 0, "Redis db")
		optTempfile  = flag.String("tempfile", "", "Temp file name")
	)
	flag.Parse()

	var resque ResquePlugin
	resque.Prefix = *optPrefix
	resque.Namespace = *optNamespace
	resque.Host = *optHost
	resque.Port = *optPort
	resque.Password = *optPassword
	resque.DB = *optDB

	err := resque.prepare()
	if err != nil {
		log.Fatalln(err)
	}

	helper := mp.NewMackerelPlugin(resque)
	helper.Tempfile = *optTempfile
	helper.Run()
}
