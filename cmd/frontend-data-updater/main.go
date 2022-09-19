package main

import (
	"eth2-exporter/cache"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"flag"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")
	flag.Parse()

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		logrus.Fatalf("error reading config file: %v", err)
	}
	utils.Config = cfg
	logrus.WithField("config", *configPath).WithField("version", version.Version).WithField("chainName", utils.Config.Chain.Config.ConfigName).Printf("starting")

	// ctx, done := context.WithTimeout(context.Background(), time.Second*30)
	// defer done()
	// db.MustInitBigtableAdmin(ctx, cfg.Bigtable.Project, cfg.Bigtable.Instance)

	// err = db.BigAdminClient.SetupBigtableCache()
	// if err != nil {
	// 	logrus.Fatalf("error setting up bigtable cache err: %v", err)
	// }

	_, err = db.InitBigtable(cfg.Bigtable.Project, cfg.Bigtable.Instance, fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID))
	if err != nil {
		logrus.Fatalf("error initializing bigtable %v", err)
	}

	db.MustInitDB(&types.DatabaseConfig{
		Username: cfg.WriterDatabase.Username,
		Password: cfg.WriterDatabase.Password,
		Name:     cfg.WriterDatabase.Name,
		Host:     cfg.WriterDatabase.Host,
		Port:     cfg.WriterDatabase.Port,
	}, &types.DatabaseConfig{
		Username: cfg.ReaderDatabase.Username,
		Password: cfg.ReaderDatabase.Password,
		Name:     cfg.ReaderDatabase.Name,
		Host:     cfg.ReaderDatabase.Host,
		Port:     cfg.ReaderDatabase.Port,
	})
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	if utils.Config.TierdCacheProvider == "redis" || len(utils.Config.RedisCacheEndpoint) != 0 {
		cache.MustInitTieredCache(utils.Config.RedisCacheEndpoint)
	} else if utils.Config.TierdCacheProvider == "bigtable" && len(utils.Config.RedisCacheEndpoint) == 0 {
		cache.MustInitTieredCacheBigtable(db.BigtableClient.GetClient(), fmt.Sprintf("%d", utils.Config.Chain.Config.DepositChainID))
	}

	if utils.Config.TierdCacheProvider != "bigtable" && utils.Config.TierdCacheProvider != "redis" {
		logrus.Fatalf("No cache provider set. Please set TierdCacheProvider (example redis, bigtable)")
	}

	logrus.Infof("initializing frontend services")
	services.Init() // Init frontend services
	logrus.Infof("frontend services initiated")

	utils.WaitForCtrlC()

	logrus.Println("exiting...")
}
