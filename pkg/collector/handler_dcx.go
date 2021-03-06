package collector

import (
	"github.com/go-kit/kit/log"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
)

const (
	DcxNamespace     = "QCE/DCX"
	DcxInstanceidKey = "directConnectConnId"
)

func init() {
	registerHandler(DcxNamespace, defaultHandlerEnabled, NewDcxHandler)
}

type dcxHandler struct {
	baseProductHandler
}

func (h *dcxHandler) GetNamespace() string {
	return DcxNamespace
}

func (h *dcxHandler) IsIncludeMetric(m *metric.TcmMetric) bool {
	return true
}

func NewDcxHandler(c *TcProductCollector, logger log.Logger) (handler productHandler, err error) {
	handler = &dcxHandler{
		baseProductHandler{
			monitorQueryKey: DcxInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
