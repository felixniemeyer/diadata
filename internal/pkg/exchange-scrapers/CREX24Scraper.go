package scrapers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/carterjones/signalr"
	"github.com/carterjones/signalr/hubs"
	"github.com/diadata-org/diadata/pkg/dia"
)

const (
	INSTRUMENTS_URL = "https://api.crex24.com/v2/public/instruments"
)

type CREX24ApiInstrument struct {
	Symbol       string `json:"symbol"`
	BaseCurrency string `json:"baseCurrency"`
}

type CREX24ApiTrade struct {
	P string
	V string
	S string
	T int64
}

type CREX24ApiTradeUpdate struct {
	I  string
	NT []CREX24ApiTrade
}

type CREX24Scraper struct {
	shutdown     chan nothing
	shutdownDone chan nothing
	error        error
	client       *signalr.Client
	exchangeName string
	chanTrades   chan *dia.Trade
	msgId        int
	pairScrapers map[string]*CREX24PairScraper
}

func NewCREX24Scraper(exchange dia.Exchange) *CREX24Scraper {
	s := &CREX24Scraper{
		shutdown:     make(chan nothing),
		shutdownDone: make(chan nothing),
		pairScrapers: make(map[string]*CREX24PairScraper),
		exchangeName: exchange.Name,
		chanTrades:   make(chan *dia.Trade),
		error:        nil,
		msgId:        1,
	}
	s.connect()
	return s
}

func (s *CREX24Scraper) Channel() chan *dia.Trade {
	return s.chanTrades
}

func (s *CREX24Scraper) Close() error {
	s.client.Close()
	close(s.chanTrades)
	return s.error
}

func (s *CREX24Scraper) NormalizePair(pair dia.Pair) (dia.Pair, error) {
	return dia.Pair{}, nil
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
	panicIfErr(err) // TODO close channel make it errorrrrr instead of panic
}

func (s *CREX24Scraper) handleMessage(msg signalr.Message) {
	for _, data := range msg.M {
		if data.M == "UpdateSource" {
			for i, argument := range data.A {
				if i == 1 { // payload argument
					payload, ok := argument.(string)
					if ok {
						var parsedUpdate CREX24ApiTradeUpdate
						log.Println(payload)
						json.NewDecoder(strings.NewReader(payload)).Decode(&parsedUpdate)
						s.sendTradesToChannel(&parsedUpdate)
					}
				}
			}
		}
	}
}

func (s *CREX24Scraper) sendTradesToChannel(update *CREX24ApiTradeUpdate) {
	ps := s.pairScrapers[update.I]
	pair := ps.Pair()
	for _, trade := range update.NT {
		price, pok := strconv.ParseFloat(trade.P, 64)
		volume, vok := strconv.ParseFloat(trade.V, 64)
		log.Println("ok?", pok, vok)
		if pok == nil && vok == nil {
			t := &dia.Trade{
				Pair:           pair.ForeignName,
				Price:          price,
				Symbol:         pair.Symbol,
				Volume:         volume,
				Time:           time.Unix(trade.T, 0),
				ForeignTradeID: "",
				Source:         dia.CREX24Exchange,
			}
			s.chanTrades <- t
		}
	}
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
	resp, err := http.Get(INSTRUMENTS_URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsedPairs []CREX24ApiInstrument
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
	log.Println(pair)
	// time.Sleep(2000 * time.Millisecond)
	msg := hubs.ClientMsg{
		H: "publicCryptoHub",
		M: "joinTradeHistory",
		A: []interface{}{pair.ForeignName},
		I: s.msgId,
	}

	err := s.client.Send(msg)
	s.msgId += 1

	if err != nil {
		return nil, err
	}

	ps := &CREX24PairScraper{
		parent: s,
		pair:   pair,
	}

	s.pairScrapers[pair.ForeignName] = ps

	return ps, nil
}

func (s *CREX24Scraper) Subscribe() error {
	return nil
}

type CREX24PairScraper struct {
	parent *CREX24Scraper
	pair   dia.Pair
}

func (ps *CREX24PairScraper) Close() error {
	// TODO prevent seg fault analog to in CREX24Scraper.ScrapPair
	err := ps.parent.client.Send(hubs.ClientMsg{
		H: "publicCryptoHub",
		M: "leaveTradeHistory",
		A: []interface{}{ps.pair.ForeignName},
		I: ps.parent.msgId,
	})
	ps.parent.msgId += 1
	if err != nil {
		return err
	}
	return nil
}

func (ps *CREX24PairScraper) Error() error {
	return ps.parent.error
}

func (ps *CREX24PairScraper) Pair() dia.Pair {
	return ps.pair
}
