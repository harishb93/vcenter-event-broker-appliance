package embedded

import (
	"io/ioutil"
	"strconv"
	"strings"
)

const (
	// map name for exposing the event router stats
	mapName = "vmware.event.router.stats"
)

// load captures the 1/5/15 load interval of a GNU/Linux system
type load struct {
	Load1  float64
	Load5  float64
	Load15 float64
}

// function that will be called by expvar to export the information from the
// structure every time the endpoint is reached
func allLoadAvg() interface{} {
	return load{
		Load1:  loadAvg(0),
		Load5:  loadAvg(1),
		Load15: loadAvg(2),
	}
}

// helper function to retrieve the load average in GNU/Linux systems
func loadAvg(position int) float64 {
	// intentionally ignoring errors to make this work under non GNU/Linux
	// systems (testing)
	data, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		//
		return 0
	}

	values := strings.Fields(string(data))
	load, err := strconv.ParseFloat(values[position], 64)

	if err != nil {
		return 0
	}

	return load
}
