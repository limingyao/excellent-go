package collector

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
)

// See https://github.com/prometheus/procfs/blob/master/proc_stat.go for details on userHZ.
const userHZ = 100

func canCollectProcess() bool {
	_, err := procfs.NewDefaultFS()
	return err == nil
}

type processExtCollector struct {
	collectFn                                  func(chan<- prometheus.Metric)
	pidFn                                      func() (int, error)
	reportErrors                               bool
	userCpuSecondsTotal, systemCpuSecondsTotal *prometheus.Desc
	readBytesTotal, writeBytesTotal            *prometheus.Desc
	readTotal, writeTotal                      *prometheus.Desc
	diskReadBytesTotal, diskWriteBytesTotal    *prometheus.Desc
	diskCancelledWriteBytesTotal               *prometheus.Desc
}

type ProcessExtCollectorOpts struct {
	PidFn        func() (int, error)
	ReportErrors bool
}

func NewProcessExtCollector(opts ProcessExtCollectorOpts) prometheus.Collector {
	c := &processExtCollector{
		reportErrors:                 opts.ReportErrors,
		userCpuSecondsTotal:          userCpuSecondsDesc,
		systemCpuSecondsTotal:        systemCpuSecondsDesc,
		readBytesTotal:               readBytesDesc,
		writeBytesTotal:              writeBytesDesc,
		readTotal:                    readDesc,
		writeTotal:                   writeDesc,
		diskReadBytesTotal:           diskReadBytesDesc,
		diskWriteBytesTotal:          diskWriteBytesDesc,
		diskCancelledWriteBytesTotal: diskCancelledWriteBytesDesc,
	}

	if opts.PidFn == nil {
		pid := os.Getpid()
		c.pidFn = func() (int, error) { return pid, nil }
	} else {
		c.pidFn = opts.PidFn
	}

	if canCollectProcess() {
		c.collectFn = c.processCollect
	} else {
		descs := []*prometheus.Desc{
			userCpuSecondsDesc, systemCpuSecondsDesc,
			readBytesDesc, writeBytesDesc,
			readDesc, writeDesc,
			diskReadBytesDesc, diskWriteBytesDesc, diskCancelledWriteBytesDesc,
		}
		c.collectFn = func(ch chan<- prometheus.Metric) {
			c.reportError(ch, descs[rand.Int()%len(descs)], fmt.Errorf("process ext metrics not supported %s/%s", runtime.GOOS, runtime.GOARCH))
		}
	}

	return c
}

func (c *processExtCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.userCpuSecondsTotal
	ch <- c.systemCpuSecondsTotal
	ch <- c.readBytesTotal
	ch <- c.writeBytesTotal
	ch <- c.readTotal
	ch <- c.writeTotal
	ch <- c.diskReadBytesTotal
	ch <- c.diskWriteBytesTotal
	ch <- c.diskCancelledWriteBytesTotal
}

func (c *processExtCollector) Collect(ch chan<- prometheus.Metric) {
	c.collectFn(ch)
}

func (c *processExtCollector) reportError(ch chan<- prometheus.Metric, desc *prometheus.Desc, err error) {
	if !c.reportErrors {
		return
	}
	if desc == nil {
		desc = prometheus.NewInvalidDesc(err)
	}
	ch <- prometheus.NewInvalidMetric(desc, err)
}

func (c *processExtCollector) processCollect(ch chan<- prometheus.Metric) {
	pid, err := c.pidFn()
	if err != nil {
		c.reportError(ch, nil, err)
		return
	}

	p, err := procfs.NewProc(pid)
	if err != nil {
		c.reportError(ch, nil, err)
		return
	}

	if stat, err := p.Stat(); err == nil {
		ch <- prometheus.MustNewConstMetric(c.userCpuSecondsTotal, prometheus.CounterValue, float64(stat.UTime)/userHZ)
		ch <- prometheus.MustNewConstMetric(c.systemCpuSecondsTotal, prometheus.CounterValue, float64(stat.STime)/userHZ)
	} else {
		c.reportError(ch, nil, err)
	}

	if io, err := p.IO(); err == nil {
		ch <- prometheus.MustNewConstMetric(c.readBytesTotal, prometheus.CounterValue, float64(io.RChar))
		ch <- prometheus.MustNewConstMetric(c.writeBytesTotal, prometheus.CounterValue, float64(io.WChar))
		ch <- prometheus.MustNewConstMetric(c.readTotal, prometheus.CounterValue, float64(io.SyscR))
		ch <- prometheus.MustNewConstMetric(c.writeTotal, prometheus.CounterValue, float64(io.SyscW))
		ch <- prometheus.MustNewConstMetric(c.diskReadBytesTotal, prometheus.CounterValue, float64(io.ReadBytes))
		ch <- prometheus.MustNewConstMetric(c.diskWriteBytesTotal, prometheus.CounterValue, float64(io.WriteBytes))
		ch <- prometheus.MustNewConstMetric(c.diskCancelledWriteBytesTotal, prometheus.CounterValue, float64(io.CancelledWriteBytes))
	} else {
		c.reportError(ch, nil, err)
	}
}

var (
	userCpuSecondsDesc = prometheus.NewDesc(
		"process_user_cpu_seconds_total",
		"Total user CPU time spent in seconds.",
		nil, nil)

	systemCpuSecondsDesc = prometheus.NewDesc(
		"process_system_cpu_seconds_total",
		"Total system CPU time spent in seconds.",
		nil, nil)

	readBytesDesc = prometheus.NewDesc(
		"io_read_bytes_total",
		"Chars read. (rchar)",
		nil, nil)

	writeBytesDesc = prometheus.NewDesc(
		"io_write_bytes_total",
		"Chars written. (wchar)",
		nil, nil)

	readDesc = prometheus.NewDesc(
		"io_read_total",
		"Read syscalls. (syscr)",
		nil, nil)

	writeDesc = prometheus.NewDesc(
		"io_write_total",
		"Write syscalls. (syscw)",
		nil, nil)

	diskReadBytesDesc = prometheus.NewDesc(
		"disk_read_bytes_total",
		"Bytes read. (read_bytes)",
		nil, nil)

	diskWriteBytesDesc = prometheus.NewDesc(
		"disk_write_bytes_total",
		"Bytes written. (write_bytes)",
		nil, nil)

	diskCancelledWriteBytesDesc = prometheus.NewDesc(
		"disk_cancelled_write_bytes_total",
		"Bytes written, but taking into account truncation. (cancelled_write_bytes)",
		nil, nil)
)
