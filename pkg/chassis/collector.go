// SPDX-FileCopyrightText: (c) Mauve Mailorder Software GmbH & Co. KG, 2022. Licensed under [MIT](LICENSE) license.
//
// SPDX-License-Identifier: MIT

package chassis

import (
	"context"
	"time"

	"github.com/MauveSoftware/ilo_exporter/pkg/chassis/power"
	"github.com/MauveSoftware/ilo_exporter/pkg/chassis/thermal"
	"github.com/MauveSoftware/ilo_exporter/pkg/client"
	"github.com/MauveSoftware/ilo_exporter/pkg/common"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	prefix = "ilo_"
)

var (
	scrapeDurationDesc = prometheus.NewDesc(prefix+"chassis_scrape_duration_second", "Scrape duration for the chassis module", []string{"host"}, nil)
)

// NewCollector returns a new collector for chassis metrics
func NewCollector(ctx context.Context, cl client.Client) prometheus.Collector {
	return &collector{
		rootCtx: ctx,
		cl:      cl,
	}
}

type collector struct {
	rootCtx context.Context
	cl      client.Client
}

// Describe implements prometheus.Collector interface
func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	power.Describe(ch)
	thermal.Describe(ch)
}

// Collect implements prometheus.Collector interface
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()

	p := "Chassis/1"

	ctx := c.rootCtx
	cc := common.NewCollectorContext(ctx, c.cl, ch)
	power.Collect(ctx, p, cc)
	thermal.Collect(ctx, p, cc)

	duration := time.Since(start).Seconds()
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration, c.cl.HostName())
}
