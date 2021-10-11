// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/cortexproject/cortex/blob/master/pkg/ingester/client/compat.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Cortex Authors.

package client

import (
	"fmt"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"

	"github.com/grafana/mimir/pkg/mimirpb"
)

// ToQueryRequest builds a QueryRequest proto.
func ToQueryRequest(from, to model.Time, matchers []*labels.Matcher) (*QueryRequest, error) {
	ms, err := toLabelMatchers(matchers)
	if err != nil {
		return nil, err
	}

	return &QueryRequest{
		StartTimestampMs: int64(from),
		EndTimestampMs:   int64(to),
		Matchers:         ms,
	}, nil
}

// FromQueryRequest unpacks a QueryRequest proto.
func FromQueryRequest(req *QueryRequest) (model.Time, model.Time, []*labels.Matcher, error) {
	matchers, err := FromLabelMatchers(req.Matchers)
	if err != nil {
		return 0, 0, nil, err
	}
	from := model.Time(req.StartTimestampMs)
	to := model.Time(req.EndTimestampMs)
	return from, to, matchers, nil
}

// ToExemplarQueryRequest builds an ExemplarQueryRequest proto.
func ToExemplarQueryRequest(from, to model.Time, matchers ...[]*labels.Matcher) (*ExemplarQueryRequest, error) {
	var reqMatchers []*LabelMatchers
	for _, m := range matchers {
		ms, err := toLabelMatchers(m)
		if err != nil {
			return nil, err
		}
		reqMatchers = append(reqMatchers, &LabelMatchers{ms})
	}

	return &ExemplarQueryRequest{
		StartTimestampMs: int64(from),
		EndTimestampMs:   int64(to),
		Matchers:         reqMatchers,
	}, nil
}

// FromExemplarQueryRequest unpacks a ExemplarQueryRequest proto.
func FromExemplarQueryRequest(req *ExemplarQueryRequest) (int64, int64, [][]*labels.Matcher, error) {
	var result [][]*labels.Matcher
	for _, m := range req.Matchers {
		matchers, err := FromLabelMatchers(m.Matchers)
		if err != nil {
			return 0, 0, nil, err
		}
		result = append(result, matchers)
	}

	return req.StartTimestampMs, req.EndTimestampMs, result, nil
}

// ToMetricsForLabelMatchersRequest builds a MetricsForLabelMatchersRequest proto
func ToMetricsForLabelMatchersRequest(from, to model.Time, matchers []*labels.Matcher) (*MetricsForLabelMatchersRequest, error) {
	ms, err := toLabelMatchers(matchers)
	if err != nil {
		return nil, err
	}

	return &MetricsForLabelMatchersRequest{
		StartTimestampMs: int64(from),
		EndTimestampMs:   int64(to),
		MatchersSet:      []*LabelMatchers{{Matchers: ms}},
	}, nil
}

// FromMetricsForLabelMatchersRequest unpacks a MetricsForLabelMatchersRequest proto
func FromMetricsForLabelMatchersRequest(req *MetricsForLabelMatchersRequest) (model.Time, model.Time, [][]*labels.Matcher, error) {
	matchersSet := make([][]*labels.Matcher, 0, len(req.MatchersSet))
	for _, matchers := range req.MatchersSet {
		matchers, err := FromLabelMatchers(matchers.Matchers)
		if err != nil {
			return 0, 0, nil, err
		}
		matchersSet = append(matchersSet, matchers)
	}
	from := model.Time(req.StartTimestampMs)
	to := model.Time(req.EndTimestampMs)
	return from, to, matchersSet, nil
}

// FromMetricsForLabelMatchersResponse unpacks a MetricsForLabelMatchersResponse proto
func FromMetricsForLabelMatchersResponse(resp *MetricsForLabelMatchersResponse) []model.Metric {
	metrics := []model.Metric{}
	for _, m := range resp.Metric {
		metrics = append(metrics, mimirpb.FromLabelAdaptersToMetric(m.Labels))
	}
	return metrics
}

// ToLabelValuesRequest builds a LabelValuesRequest proto
func ToLabelValuesRequest(labelName model.LabelName, from, to model.Time, matchers []*labels.Matcher) (*LabelValuesRequest, error) {
	ms, err := toLabelMatchers(matchers)
	if err != nil {
		return nil, err
	}

	return &LabelValuesRequest{
		LabelName:        string(labelName),
		StartTimestampMs: int64(from),
		EndTimestampMs:   int64(to),
		Matchers:         &LabelMatchers{Matchers: ms},
	}, nil
}

func ToLabelValuesCardinalityRequest(labelNames []model.LabelName, matchers []*labels.Matcher) (*LabelValuesCardinalityRequest, error) {
	matchersProto, err := toLabelMatchers(matchers)
	if err != nil {
		return nil, err
	}
	labelNamesStr := make([]string, 0, len(labelNames))
	for _, labelName := range labelNames {
		labelNamesStr = append(labelNamesStr, string(labelName))
	}
	return &LabelValuesCardinalityRequest{LabelNames: labelNamesStr, Matchers: matchersProto}, nil
}

// FromLabelValuesRequest unpacks a LabelValuesRequest proto
func FromLabelValuesRequest(req *LabelValuesRequest) (string, int64, int64, []*labels.Matcher, error) {
	var err error
	var matchers []*labels.Matcher

	if req.Matchers != nil {
		matchers, err = FromLabelMatchers(req.Matchers.Matchers)
		if err != nil {
			return "", 0, 0, nil, err
		}
	}

	return req.LabelName, req.StartTimestampMs, req.EndTimestampMs, matchers, nil
}

// ToLabelNamesRequest builds a LabelNamesRequest proto
func ToLabelNamesRequest(from, to model.Time, matchers []*labels.Matcher) (*LabelNamesRequest, error) {
	ms, err := toLabelMatchers(matchers)
	if err != nil {
		return nil, err
	}

	return &LabelNamesRequest{
		StartTimestampMs: int64(from),
		EndTimestampMs:   int64(to),
		Matchers:         &LabelMatchers{Matchers: ms},
	}, nil
}

// FromLabelNamesRequest unpacks a LabelNamesRequest proto
func FromLabelNamesRequest(req *LabelNamesRequest) (int64, int64, []*labels.Matcher, error) {
	var err error
	var matchers []*labels.Matcher

	if req.Matchers != nil {
		matchers, err = FromLabelMatchers(req.Matchers.Matchers)
		if err != nil {
			return 0, 0, nil, err
		}
	}

	return req.StartTimestampMs, req.EndTimestampMs, matchers, nil
}

func toLabelMatchers(matchers []*labels.Matcher) ([]*LabelMatcher, error) {
	result := make([]*LabelMatcher, 0, len(matchers))
	for _, matcher := range matchers {
		var mType MatchType
		switch matcher.Type {
		case labels.MatchEqual:
			mType = EQUAL
		case labels.MatchNotEqual:
			mType = NOT_EQUAL
		case labels.MatchRegexp:
			mType = REGEX_MATCH
		case labels.MatchNotRegexp:
			mType = REGEX_NO_MATCH
		default:
			return nil, fmt.Errorf("invalid matcher type")
		}
		result = append(result, &LabelMatcher{
			Type:  mType,
			Name:  matcher.Name,
			Value: matcher.Value,
		})
	}
	return result, nil
}

func FromLabelMatchers(matchers []*LabelMatcher) ([]*labels.Matcher, error) {
	result := make([]*labels.Matcher, 0, len(matchers))
	for _, matcher := range matchers {
		var mtype labels.MatchType
		switch matcher.Type {
		case EQUAL:
			mtype = labels.MatchEqual
		case NOT_EQUAL:
			mtype = labels.MatchNotEqual
		case REGEX_MATCH:
			mtype = labels.MatchRegexp
		case REGEX_NO_MATCH:
			mtype = labels.MatchNotRegexp
		default:
			return nil, fmt.Errorf("invalid matcher type")
		}
		matcher, err := labels.NewMatcher(mtype, matcher.Name, matcher.Value)
		if err != nil {
			return nil, err
		}
		result = append(result, matcher)
	}
	return result, nil
}

// FastFingerprint runs the same algorithm as Prometheus labelSetToFastFingerprint()
func FastFingerprint(ls []mimirpb.LabelAdapter) model.Fingerprint {
	if len(ls) == 0 {
		return model.Metric(nil).FastFingerprint()
	}

	var result uint64
	for _, l := range ls {
		sum := hashNew()
		sum = hashAdd(sum, l.Name)
		sum = hashAddByte(sum, model.SeparatorByte)
		sum = hashAdd(sum, l.Value)
		result ^= sum
	}
	return model.Fingerprint(result)
}

// Fingerprint runs the same algorithm as Prometheus labelSetToFingerprint()
func Fingerprint(labels labels.Labels) model.Fingerprint {
	sum := hashNew()
	for _, label := range labels {
		sum = hashAddString(sum, label.Name)
		sum = hashAddByte(sum, model.SeparatorByte)
		sum = hashAddString(sum, label.Value)
		sum = hashAddByte(sum, model.SeparatorByte)
	}
	return model.Fingerprint(sum)
}

// LabelsToKeyString is used to form a string to be used as
// the hashKey. Don't print, use l.String() for printing.
func LabelsToKeyString(l labels.Labels) string {
	// We are allocating 1024, even though most series are less than 600b long.
	// But this is not an issue as this function is being inlined when called in a loop
	// and buffer allocated is a static buffer and not a dynamic buffer on the heap.
	b := make([]byte, 0, 1024)
	return string(l.Bytes(b))
}
