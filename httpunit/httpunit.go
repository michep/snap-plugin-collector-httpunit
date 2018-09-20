package httpunit

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/StackExchange/httpunit"
	"github.com/ghodss/yaml"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	PluginName    = "httpunit"
	PluginVersion = 1
	PluginVedor   = "mfms"
)

type Plugin struct {
	initialized bool
	plans       httpunit.Plans
}

func NewCollector() *Plugin {
	return &Plugin{initialized: false}
}

func (p *Plugin) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewStringRule([]string{PluginVedor, PluginName}, "setfile", true)
	return *policy, nil
}

func (p *Plugin) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}
	for _, name := range []string{"health", "time"} {
		metrics = append(metrics, plugin.Metric{
			Namespace: createNamespace(name),
		})
	}
	return metrics, nil
}

func (p *Plugin) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}

	if !p.initialized {
		setfile, err := mts[0].Config.GetString("setfile")
		if err != nil {
			return nil, err
		}

		err = p.getPlansFromConfig(setfile)
		if err != nil {
			return nil, err
		}
		p.initialized = true
	}

	ts := time.Now()

	resCh, _, err := p.plans.Test("", false, []string{}, []string{})
	if err != nil {
		return nil, err
	}

	var m plugin.Metric
	var healthdata, timedata int
	var ip string
	type resStruct struct {
		data  interface{}
		label string
		ip    string
	}
	resMap := map[string][]resStruct{}
	for r := range resCh {
		if r.Result.GotCode && r.Result.GotRegex && r.Result.GotText {
			healthdata = 0
			timedata = int(r.Result.TimeTotal.Seconds() * 1000)
		} else {
			healthdata = 1
			timedata = 0
		}
		ip = r.Case.IP.String()
		if ip == "<nil>" {
			ip = "NA"
		}
		resMap["time"] = append(resMap["time"], resStruct{data: timedata, label: r.Plan.Label, ip: ip})
		resMap["health"] = append(resMap["health"], resStruct{data: healthdata, label: r.Plan.Label, ip: ip})
	}
	for _, metric := range mts {
		name := metric.Namespace[4].Value
		for _, res := range resMap[name] {
			m = plugin.Metric{
				Namespace: createNamespace(name),
				Timestamp: ts,
				Data:      res.data,
			}
			m.Namespace[2].Value = res.label
			m.Namespace[3].Value = res.ip
			metrics = append(metrics, m)
		}
	}
	return metrics, nil
}

func (p *Plugin) getPlansFromConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &p.plans)
	if err != nil {
		return fmt.Errorf("Error parsing YAML httpunit setfile: %v", err)
	}
	return nil
}

func createNamespace(name string) plugin.Namespace {
	namespace := plugin.NewNamespace(PluginVedor, PluginName)
	namespace = namespace.AddDynamicElement("resource", "resource name")
	namespace = namespace.AddDynamicElement("ip", "ip address")
	namespace = namespace.AddStaticElement(name)
	return namespace
}
