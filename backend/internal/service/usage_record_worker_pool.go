package service

import (
	"github.com/Aias00/cloudbase/internal/config"
	"github.com/Aias00/cloudbase/internal/worker"
)

type UsageRecordTask = worker.UsageRecordTask
type UsageRecordSubmitMode = worker.UsageRecordSubmitMode
type UsageRecordWorkerPoolOptions = worker.UsageRecordWorkerPoolOptions
type UsageRecordWorkerPoolStats = worker.UsageRecordWorkerPoolStats
type UsageRecordWorkerPool = worker.UsageRecordWorkerPool

const (
	UsageRecordSubmitModeEnqueued = worker.UsageRecordSubmitModeEnqueued
	UsageRecordSubmitModeDropped  = worker.UsageRecordSubmitModeDropped
	UsageRecordSubmitModeSync     = worker.UsageRecordSubmitModeSync
)

func NewUsageRecordWorkerPool(cfg *config.Config) *UsageRecordWorkerPool {
	return worker.NewUsageRecordWorkerPool(cfg)
}

func NewUsageRecordWorkerPoolWithOptions(opts UsageRecordWorkerPoolOptions) *UsageRecordWorkerPool {
	return worker.NewUsageRecordWorkerPoolWithOptions(opts)
}
