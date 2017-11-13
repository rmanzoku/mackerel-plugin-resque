package mpresque

import (
	"flag"

	mp "github.com/mackerelio/go-mackerel-plugin-helper"
)

var graphdef = map[string]mp.Graphs{
	"queue": {
		Label: "Resque Queue",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "queue", Label: "Queue", Diff: false},
		},
	},
}

// ResquePlugin mackerel plugin for Resque
type ResquePlugin struct {
	Prefix    string
	KeyPrefix string
	Host      string
	Port      string
}

// FetchMetrics interface for mackerelplugin
func (r ResquePlugin) FetchMetrics() (map[string]interface{}, error) {
	ret := make(map[string]interface{})

	ret["queue"] = 1.0
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
		optKeyPrefix = flag.String("prefix", "resque", "Redis key prefix")
		optHost      = flag.String("host", "127.0.0.1", "The bind url to use for the redis server")
		optPort      = flag.String("port", "6379", "The bind port to use for the redis server")
		optTempfile  = flag.String("tempfile", "", "Temp file name")
	)
	flag.Parse()

	var resque ResquePlugin
	resque.Prefix = *optPrefix
	resque.KeyPrefix = *optKeyPrefix
	resque.Host = *optHost
	resque.Port = *optPort

	helper := mp.NewMackerelPlugin(resque)
	helper.Tempfile = *optTempfile
	helper.Run()
}
