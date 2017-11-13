package mpresque

import (
	"flag"
	"fmt"

	"github.com/go-redis/redis"
	mp "github.com/mackerelio/go-mackerel-plugin-helper"
)

var graphdef = map[string]mp.Graphs{
	"queue.#": {
		Label: "Resque Queue",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "pending", Label: "Pending", Diff: false, Stacked: true},
		},
	},
	"worker": {
		Label: "Resque Worker",
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
}

// FetchMetrics interface for mackerelplugin
func (r ResquePlugin) FetchMetrics() (map[string]interface{}, error) {
	ret := make(map[string]interface{})

	addr := fmt.Sprintf("%s:%s", r.Host, r.Port)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: r.Password,
		DB:       r.DB,
		Network:  "tcp",
	})

	queuesKey := fmt.Sprintf("%s:%s", r.Namespace, "queues")
	queues, err := redisClient.SMembers(queuesKey).Result()
	if err != nil {
		panic(err)
	}

	for _, q := range queues {

		qKey := fmt.Sprintf("%s:%s:%s", r.Namespace, "queue", q)
		qlen, err := redisClient.LLen(qKey).Result()
		if err != nil {
			panic(err)
		}

		ret["queue."+q+".pending"] = float64(qlen)

	}

	workerKey := fmt.Sprintf("%s:%s", r.Namespace, "workers")
	workerProcesses, err := redisClient.SCard(workerKey).Result()
	if err != nil {
		panic(err)
	}
	ret["processes"] = float64(workerProcesses)

	failedKey := fmt.Sprintf("%s:%s", r.Namespace, "stat:failed")
	failedCount, err := redisClient.Get(failedKey).Float64()
	if err != nil {
		panic(err)
	}
	ret["failed"] = failedCount

	processedKey := fmt.Sprintf("%s:%s", r.Namespace, "stat:processed")
	processedCount, err := redisClient.Get(processedKey).Float64()
	if err != nil {
		panic(err)
	}
	ret["processed"] = processedCount

	return ret, nil

}

// GraphDefinition interface for mackerelplugin
func (r ResquePlugin) GraphDefinition() map[string]mp.Graphs {
	graphdef := graphdef
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

	helper := mp.NewMackerelPlugin(resque)
	helper.Tempfile = *optTempfile
	helper.Run()
}
