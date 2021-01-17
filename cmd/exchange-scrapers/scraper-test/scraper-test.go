package main

import (
	"flag"
	"sync"
	"time"

	scrapers "github.com/diadata-org/diadata/internal/pkg/exchange-scrapers"
	"github.com/diadata-org/diadata/pkg/dia"
	"github.com/diadata-org/diadata/pkg/dia/helpers/configCollectors"
	models "github.com/diadata-org/diadata/pkg/model"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
}

func handleTrades(es *scrapers.APIScraper, wg *sync.WaitGroup) {
	close := time.NewTimer(time.Second * 10)
	c := (*es).Channel()
	for {
		select {
		case <-close.C:
			(*es).Close()
			return
		case t, ok := <-c:
			if !ok {
				wg.Done()
				log.Error("handleTrades")
				return
			}
			log.Println(t)
		}
	}
}

var (
	exchange         = flag.String("exchange", "", "which exchange")
	onePairPerSymbol = flag.Bool("onePairPerSymbol", false, "one Pair max Per Symbol ?")
)

func init() {
	flag.Parse()
	if *exchange == "" {
		flag.Usage()
		log.Println(dia.Exchanges())
		log.Fatal("exchange is required")
	}
}

// main manages all PairScrapers and handles incoming trade information
func main() {
	ds, err := models.NewRedisDataStore()
	if err != nil {
		log.Errorln("NewDataStore:", err)
	}
	pairsExchange, err := ds.GetAvailablePairsForExchange(*exchange)
	log.Info("available pairs:", len(pairsExchange))

	if err != nil || len(pairsExchange) == 0 {
		log.Error("error on GetAvailablePairsForExchange", err)
		cc := configCollectors.NewConfigCollectors(*exchange)
		pairsExchange = cc.AllPairs()
	}

	configApi, err := dia.GetConfig(*exchange)
	if err != nil {
		log.Warning("no config for exchange's api ", err)
	}
	es := scrapers.NewAPIScraper(*exchange, configApi.ApiKey, configApi.SecretKey)

	wg := sync.WaitGroup{}

	pairs := make(map[string]string)

	for _, configPair := range pairsExchange {
		dontAddPair := false
		if *onePairPerSymbol {
			_, dontAddPair = pairs[configPair.Symbol]
			pairs[configPair.Symbol] = configPair.Symbol
		}
		if dontAddPair {
			log.Println("Skipping pair:", configPair.Symbol, configPair.ForeignName, "on exchange", *exchange)
		} else {
			log.Println("Adding pair:", configPair.Symbol, configPair.ForeignName, "on exchange", *exchange)
			_, err := es.ScrapePair(dia.Pair{
				Symbol:      configPair.Symbol,
				ForeignName: configPair.ForeignName})
			if err != nil {
				log.Println(err)
			} else {
				wg.Add(1)
			}
		}
		defer wg.Wait()
	}
	go handleTrades(&es, &wg)
}
