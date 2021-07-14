package main

import (
	"github.com/goharbor/harbor/src/common/models"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"

	"github.com/goharbor/harbor/src/common/dao"
	commonthttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/exporter"
)

func main() {
	viper.SetEnvPrefix("harbor")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	dbCfg := &models.Database{
		Type: "postgresql",
		PostGreSQL: &models.PostGreSQL{
			Host:         viper.GetString("database.host"),
			Port:         viper.GetInt("database.port"),
			Username:     viper.GetString("database.username"),
			Password:     viper.GetString("database.password"),
			Database:     viper.GetString("database.dbname"),
			SSLMode:      viper.GetString("database.sslmode"),
			MaxIdleConns: viper.GetInt("database.max_idle_conns"),
			MaxOpenConns: viper.GetInt("database.max_open_conns"),
		},
	}
	if err := dao.InitDatabase(dbCfg); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	exporter.InitHarborClient(&exporter.HarborClient{
		HarborScheme: viper.GetString("service.scheme"),
		HarborHost:   viper.GetString("service.host"),
		HarborPort:   viper.GetInt("service.port"),
		Client: &http.Client{
			Transport: commonthttp.GetHTTPTransport(commonthttp.SecureTransport),
		},
	})

	exporter.InitBackendWorker(&exporter.RedisPoolConfig{
		URL:               viper.GetString("redis.url"),
		Namespace:         viper.GetString("redis.namespace"),
		IdleTimeoutSecond: viper.GetInt("redis.timeout"),
	})

	exporterOpt := &exporter.Opt{
		Port:                   viper.GetInt("exporter.port"),
		MetricsPath:            viper.GetString("exporter.metrics_path"),
		ExporterMetricsEnabled: viper.GetBool("exporter.metrics_enabled"),
		MaxRequests:            viper.GetInt("exporter.max_requests"),
		TLSEnabled:             viper.GetBool("exporter.tls_enabled"),
		Certificate:            viper.GetString("exporter.tls_cert"),
		Key:                    viper.GetString("exporter.tls_key"),
		CacheDuration:          viper.GetInt64("exporter.cache_time"),
		CacheCleanInterval:     viper.GetInt64("exporter.cache_clean_interval"),
	}
	harborExporter := exporter.NewExporter(exporterOpt)
	log.Infof("Starting harbor_exporter with port=%v path=%v metrics=%v max_request=%v tls=%v cert=%v key=%v cache_time=%v clean_internal=%v",
		exporterOpt.Port,
		exporterOpt.MetricsPath,
		exporterOpt.ExporterMetricsEnabled,
		exporterOpt.MaxRequests,
		exporterOpt.TLSEnabled,
		exporterOpt.Certificate,
		exporterOpt.Key,
		exporterOpt.CacheDuration,
		exporterOpt.CacheCleanInterval,
	)
	prometheus.MustRegister(harborExporter)
	if err := harborExporter.ListenAndServe(); err != nil {
		log.Errorf("Error starting Harbor expoter %s", err)
		os.Exit(1)
	}
}
