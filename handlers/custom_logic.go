package handlers

//go:generate $GOPATH/bin/mockgen -source=$GOFILE -destination=$PWD/mocks/$GOFILE -package=mocks

import (
	"io"
	"net/http"
	"time"

	"github.com/gracew/widget-proxy/metrics"
	"github.com/pkg/errors"
)

type CustomLogicExecutor interface {
	Execute(reader io.Reader, when string, operation string) (*http.Response, error)
}

type RemoteCustomLogicExecutor struct {
	URL string
}

func (c RemoteCustomLogicExecutor) Execute(reader io.Reader, when string, operation string) (*http.Response, error) {
	start := time.Now()
	res, err := http.Post(c.URL+when+operation, "application/json", reader)
	if err != nil {
		metrics.CustomLogicErrors.WithLabelValues(operation, when).Inc()
		return nil, errors.Wrap(err, "request to custom logic endpoint failed")
	}
	end := time.Now()
	metrics.CustomLogicSummary.WithLabelValues(operation, when).Observe(end.Sub(start).Seconds())
	return res, nil
}
