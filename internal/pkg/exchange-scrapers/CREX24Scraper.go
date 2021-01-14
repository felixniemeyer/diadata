package scrapers

import (
	"encoding/json"
	"net/http"

	"github.com/carterjones/signalr"
	"github.com/diadata-org/diadata/pkg/dia"
)

const (
	INSTRUMENTS_URL = "https://api.crex24.com/v2/public/instruments"
)

type CREX24Instrument struct {
	Symbol       string `json:"symbol"`
	BaseCurrency string `json:"baseCurrency"`
	// 	  QuoteCurrency string `json:"quoteCurrency"`
	//    feeCurrency		string
	//    feeSchedule		string
	//    tickSize		tick
	//    minPrice
	//    maxPrice
	//    volumeIncrement
	//    minVolume
	//    maxVolume
	//    minQuoteVolume
	//    maxQuoteVolume
	//    supportedOrderTypes
	//      limit
	//    ],
	//    state
}

type CREX24Scraper struct {
	shutdown     chan nothing
	shutdownDone chan nothing
	error        error
	client       *signalr.Client
	exchangeName string
	chanTrades   chan *dia.Trade
}

func NewCREX24Scraper(exchange dia.Exchange) *CREX24Scraper {
	s := &CREX24Scraper{
		shutdown:     make(chan nothing),
		shutdownDone: make(chan nothing),
		/// pairScrapers: mak
		exchangeName: exchange.Name,
		chanTrades:   make(chan *dia.Trade),
		error:        nil,
	}
	go s.connect()
	return s
}

func (s *CREX24Scraper) Channel() chan *dia.Trade {
	return s.chanTrades
}

func (s *CREX24Scraper) Close() error {
	close(s.chanTrades)
	return s.error
}

func (s *CREX24Scraper) NormalizePair(pair dia.Pair) (dia.Pair, error) {
	return dia.Pair{}, nil
}

func (s *CREX24Scraper) mainLoop() {
	s.connect()
	for {
		select {
		case <-s.shutdown:
			s.cleanup(nil)
			return
		}
	}
}

func (s *CREX24Scraper) connect() {
	s.client = signalr.New(
		"api.crex24.com",
		"1.5",
		"/signalr",
		`[{"name": "publicCryptoHub"}]`,
		nil,
	)
	msgHandler := s.handleMessage
	panicIfErr := func(err error) {
		if err != nil {
			log.Panic(err)
		}
	}
	err := s.client.Run(msgHandler, panicIfErr)
	panicIfErr(err)
}

func (s *CREX24Scraper) handleMessage(msg signalr.Message) {
	log.Println(msg)
	// TODO parse msg and s.chanTrades <- parsed
}

func (s *CREX24Scraper) setup() {
}

func (s *CREX24Scraper) cleanup(err error) {
	s.client.Close()
	if err != nil {
		s.error = err
	}
	//close websocket connections
}

func (s *CREX24Scraper) FetchAvailablePairs() (pairs []dia.Pair, err error) {
	log.Println("reached this here")
	resp, err := http.Get(INSTRUMENTS_URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsedPairs []CREX24Instrument
	err = json.NewDecoder(resp.Body).Decode(&parsedPairs)
	if err != nil {
		return nil, err
	}

	log.Println(parsedPairs) // TODO: remove

	var results = make([]dia.Pair, len(parsedPairs))
	for i := 0; i < len(parsedPairs); i++ {
		results[i] = dia.Pair{
			Symbol:      parsedPairs[i].BaseCurrency,
			ForeignName: parsedPairs[i].Symbol,
			Ignore:      false,
			Exchange:    s.exchangeName,
		}
	}

	return results, nil
}

func (s *CREX24Scraper) ScrapePair(pair dia.Pair) (PairScraper, error) {
	return nil, nil
}
