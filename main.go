package main

import (
	"flag"
	"fmt"
	_ "time/tzdata"

	"github.com/masx200/speedtest-go/config"
	"github.com/masx200/speedtest-go/database"
	"github.com/masx200/speedtest-go/results"
	"github.com/masx200/speedtest-go/web"

	_ "github.com/breml/rootcerts"
	log "github.com/sirupsen/logrus"
)

var (
	optConfig = flag.String("c", "", "config file to be used, defaults to settings.toml in the same directory")
)

func main() {
	flag.Parse()
	conf := config.Load(*optConfig)
	printConfig(conf)
	web.SetServerLocation(&conf)
	results.Initialize(&conf)
	database.SetDBInfo(&conf)
	log.Fatal(web.ListenAndServe(&conf))
}
func printConfig(config config.Config) {
	fmt.Printf("Config {\n")
	fmt.Printf("BindAddress: %s\n", config.BindAddress)
	fmt.Printf("Port: %s\n", config.Port)
	fmt.Printf("ProxyProtocolPort: %s\n", config.ProxyProtocolPort)
	fmt.Printf("ServerLat: %f\n", config.ServerLat)
	fmt.Printf("ServerLng: %f\n", config.ServerLng)
	fmt.Printf("IPInfoAPIKey: %s\n", config.IPInfoAPIKey)
	fmt.Printf("StatsPassword: %s\n", config.StatsPassword)
	fmt.Printf("RedactIP: %t\n", config.RedactIP)
	fmt.Printf("AssetsPath: %s\n", config.AssetsPath)
	fmt.Printf("DatabaseType: %s\n", config.DatabaseType)
	fmt.Printf("DatabaseHostname: %s\n", config.DatabaseHostname)
	fmt.Printf("DatabaseName: %s\n", config.DatabaseName)
	fmt.Printf("DatabaseUsername: %s\n", config.DatabaseUsername)
	fmt.Printf("DatabasePassword: %s\n", config.DatabasePassword)
	fmt.Printf("DatabaseFile: %s\n", config.DatabaseFile)
	fmt.Printf("EnableHTTP3: %t\n", config.EnableHTTP3)
	fmt.Printf("EnableHTTP2: %t\n", config.EnableHTTP2)
	fmt.Printf("EnableTLS: %t\n", config.EnableTLS)
	fmt.Printf("TLSCertFile: %s\n", config.TLSCertFile)
	fmt.Printf("TLSKeyFile: %s\n", config.TLSKeyFile)
	fmt.Printf("} Config\n")
}
