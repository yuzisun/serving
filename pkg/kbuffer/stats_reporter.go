/*
Copyright 2018 Google Inc. All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kbuffer

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// Measurement represents the type of the autoscaler metric to be reported
type Measurement int

const (
	//RequestCountM is the request count when KBuffer proxy the request
	RequestCountM = iota

	// ResponseTimeInMsecM is the response time in millisecond
	ResponseTimeInMsecM
)

var (
	measurements = []*stats.Float64Measure{
		RequestCountM: stats.Float64(
			"revision_request_count",
			"The number of requests that are routed to KBuffer",
			stats.UnitNone),
		ResponseTimeInMsecM: stats.Float64(
			"response_time_msec",
			"The response time in millisecond",
			stats.UnitNone),
	}
)

// StatsReporter defines the interface for sending activator metrics
type StatsReporter interface {
	ReportRequestCount(ns, service, config, rev string, responseCode, numTries int, v float64) error
	ReportResponseTime(ns, service, config, rev string, responseCode int, d time.Duration) error
}

// Reporter holds cached metric objects to report autoscaler metrics
type Reporter struct {
	initialized     bool
	namespaceTagKey tag.Key
	serviceTagKey   tag.Key
	configTagKey    tag.Key
	revisionTagKey  tag.Key
	responseCodeKey tag.Key
	numTriesKey     tag.Key
}

// NewStatsReporter creates a reporter that collects and reports activator metrics
func NewStatsReporter() (*Reporter, error) {

	var r = &Reporter{}

	// Create the tag keys that will be used to add tags to our measurements.
	nsTag, err := tag.NewKey("destination_namespace")
	if err != nil {
		return nil, err
	}
	r.namespaceTagKey = nsTag
	serviceTag, err := tag.NewKey("destination_service")
	if err != nil {
		return nil, err
	}
	r.serviceTagKey = serviceTag
	configTag, err := tag.NewKey("destination_configuration")
	if err != nil {
		return nil, err
	}
	r.configTagKey = configTag
	revTag, err := tag.NewKey("destination_revision")
	if err != nil {
		return nil, err
	}
	r.revisionTagKey = revTag
	responseCodeTag, err := tag.NewKey("response_code")
	if err != nil {
		return nil, err
	}
	r.responseCodeKey = responseCodeTag
	numTriesTag, err := tag.NewKey("num_tries")
	if err != nil {
		return nil, err
	}
	r.numTriesKey = numTriesTag
	// Create view to see our measurements.
	err = view.Register(
		&view.View{
			Description: "The number of requests that are routed to KBuffer",
			Measure:     measurements[RequestCountM],
			Aggregation: view.Sum(),
			TagKeys:     []tag.Key{r.namespaceTagKey, r.serviceTagKey, r.configTagKey, r.revisionTagKey, r.responseCodeKey, r.numTriesKey},
		},
		&view.View{
			Description: "The response time in millisecond",
			Measure:     measurements[ResponseTimeInMsecM],
			Aggregation: view.Distribution(1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000, 11000, 12000, 13000, 14000, 15000),
			TagKeys:     []tag.Key{r.namespaceTagKey, r.serviceTagKey, r.configTagKey, r.revisionTagKey, r.responseCodeKey},
		},
	)
	if err != nil {
		return nil, err
	}

	r.initialized = true
	return r, nil
}

// ReportRequestCount captures request count metric with value v.
func (r *Reporter) ReportRequestCount(ns, service, config, rev string, responseCode, numTries int, v float64) error {
	if !r.initialized {
		return errors.New("StatsReporter is not initialized yet")
	}

	ctx, err := tag.New(
		context.Background(),
		tag.Insert(r.namespaceTagKey, ns),
		tag.Insert(r.serviceTagKey, service),
		tag.Insert(r.configTagKey, config),
		tag.Insert(r.revisionTagKey, rev),
		tag.Insert(r.responseCodeKey, strconv.Itoa(responseCode)),
		tag.Insert(r.numTriesKey, strconv.Itoa(numTries)))
	if err != nil {
		return err
	}

	stats.Record(ctx, measurements[RequestCountM].M(v))
	return nil
}

// ReportResponseTime captures response time requests
func (r *Reporter) ReportResponseTime(ns, service, config, rev string, responseCode int, d time.Duration) error {
	if !r.initialized {
		return errors.New("StatsReporter is not initialized yet")
	}

	ctx, err := tag.New(
		context.Background(),
		tag.Insert(r.namespaceTagKey, ns),
		tag.Insert(r.serviceTagKey, service),
		tag.Insert(r.configTagKey, config),
		tag.Insert(r.revisionTagKey, rev),
		tag.Insert(r.responseCodeKey, strconv.Itoa(responseCode)))
	if err != nil {
		return err
	}

	// convert time.Duration in nanoseconds to milliseconds
	stats.Record(ctx, measurements[ResponseTimeInMsecM].M(float64(d/time.Millisecond)))
	return nil
}
