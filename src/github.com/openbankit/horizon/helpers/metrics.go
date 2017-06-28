package helpers

import (
	"github.com/rcrowley/go-metrics"
	"time"
)

func PopulateTimer(values map[string]interface{}, timer metrics.Timer) {
	ps := timer.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
	values["count"] = timer.Count()
	values["min"] = time.Duration(timer.Min()).String()
	values["max"] = time.Duration(timer.Max()).String()
	values["mean"] = time.Duration(timer.Mean()).String()
	values["stddev"] = timer.StdDev()
	values["median"] = time.Duration(ps[0]).String()
	values["75%"] = time.Duration(ps[1]).String()
	values["95%"] = time.Duration(ps[2]).String()
	values["99%"] = time.Duration(ps[3]).String()
	values["99.9%"] = time.Duration(ps[4]).String()
	values["1m.rate"] = timer.Rate1()
	values["5m.rate"] = timer.Rate5()
	values["15m.rate"] = timer.Rate15()
	values["mean.rate"] = timer.RateMean()
}
