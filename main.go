package main

import (
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/michep/snap-plugin-collector-httpunit/httpunit"
)

func main() {
	plugin.StartCollector(httpunit.NewCollector(), httpunit.PluginName, httpunit.PluginVersion, plugin.RoutingStrategy(plugin.StickyRouter))
}
