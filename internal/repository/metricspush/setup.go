package metricspush

import (
	"github.com/StephanHCB/go-mailer-service/internal/repository/configuration"
	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	"github.com/rs/zerolog/log"
	"time"
)

func SetupPrometheusPushSink() error {
	pushInterval := 10 * time.Second
	address := configuration.MetricsPushAddress()
	name := configuration.MetricsPushSinkName()
	log.Info().Msgf("setting up prometheus metrics push sink %s on %s with push interval %v", name, address, pushInterval)

	sink, err := prometheus.NewPrometheusPushSink(address, pushInterval, name)
	if err != nil {
		return err
	}
	metricsConf := metrics.DefaultConfig(configuration.ServiceName())
	// metricsConf.HostName = "localhost"
	// metricsConf.EnableHostnameLabel = true
	_, err = metrics.NewGlobal(metricsConf, sink)
	return err
}

func SetupInMemorySink() error {
	log.Info().Msg("setting up in memory metrics push sink")
	sink := metrics.NewInmemSink(10*time.Millisecond, 50*time.Millisecond)
	metricsConf := metrics.DefaultConfig(configuration.ServiceName())
	_, err := metrics.NewGlobal(metricsConf, sink)
	return err
}

func Setup() {
	var err error
	var sinktype string
	if configuration.EnableMetricsPush() {
		err = SetupPrometheusPushSink()
		sinktype = "prometheus"
	} else {
		err = SetupInMemorySink()
		sinktype = "inmemory"
	}
	if err != nil {
		log.Fatal().Err(err).Msg("setting up metrics push sink failed for type " + sinktype)
	}
}
