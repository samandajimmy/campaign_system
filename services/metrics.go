package services

import (
	"gade/srv-gade-point/metrics"
	"gade/srv-gade-point/models"
)

// MetricService is to handle all metrics service
type MetricService struct {
	metricUsecase metrics.UseCase
}

var some MetricService

// NewMetricHandler is to return a metric service struct
func NewMetricHandler(mu metrics.UseCase) {
	some = MetricService{
		metricUsecase: mu,
	}
}

// AddMetric is to add metric to db
func AddMetric(job string) error {

	err := some.metricUsecase.AddMetric(job)

	if err != nil {
		return models.ErrCreateMetric
	}

	return err
}
