package dia

import (
	"os/user"
	"strings"
	"time"

	"github.com/tkanos/gonfig"
)

const (
	BlockSizeSeconds  = 120
	BalancerExchange  = "Balancer"
	BancorExchange    = "Bancor"
	BinanceExchange   = "Binance"
	BitBayExchange    = "BitBay"
	BitfinexExchange  = "Bitfinex"
	BitMaxExchange    = "Bitmax"
	BittrexExchange   = "Bittrex"
	CoinBaseExchange  = "CoinBase"
	CREX24Exchange    = "CREX24"
	CurveFIExchange   = "Curvefi"
	DforceExchange    = "Dforce"
	FilterKing        = "MA120"
	GateIOExchange    = "GateIO"
	GnosisExchange    = "Gnosis"
	HitBTCExchange    = "HitBTC"
	HuobiExchange     = "Huobi"
	KrakenExchange    = "Kraken"
	KuCoinExchange    = "KuCoin"
	KyberExchange     = "Kyber"
	LBankExchange     = "LBank"
	LoopringExchange  = "Loopring"
	MakerExchange     = "Maker"
	OKExExchange      = "OKEx"
	PanCakeSwap       = "PanCakeSwap"
	QuoineExchange    = "Quoine"
	SimexExchange     = "Simex"
	SushiSwapExchange = "SushiSwap"
	UniswapExchange   = "Uniswap"
	UnknownExchange   = "Unknown"
	ZBExchange        = "ZB"
	ZeroxExchange     = "0x"
)

const (
	Bitcoin  = "Bitcoin"
	Ethereum = "Ethereum"
)

func Exchanges() []string {
	return []string{
		/*
			KuCoinExchange,
			UniswapExchange,
			BalancerExchange,
			MakerExchange,
			GnosisExchange,
			CurveFIExchange,
			BinanceExchange,
			BitfinexExchange,
			BittrexExchange,
			CoinBaseExchange,
			GateIOExchange,
			HitBTCExchange,
			HuobiExchange,
		*/
		KrakenExchange,
		CREX24Exchange,
		/*
			LBankExchange,
			OKExExchange,
			QuoineExchange,
			SimexExchange,
			ZBExchange,
			BancorExchange,
			UnknownExchange,
			LoopringExchange,
			SushiSwapExchange,
			DforceExchange,
			ZeroxExchange,
			KyberExchange,
			BitMaxExchange,
			PanCakeSwap,
		*/
	}
}

type ConfigApi struct {
	ApiKey    string
	SecretKey string
}

type ConfigConnector struct {
	Coins []Pair
}

type BlockChain struct {
	Name                  string
	GenesisDate           time.Time
	NativeToken           string
	VerificationMechanism VerificationMechanism
}

type VerificationMechanism string

const (
	PROOF_OF_STAKE VerificationMechanism = "pos"
)

func GetConfig(exchange string) (*ConfigApi, error) {
	var configApi ConfigApi
	usr, _ := user.Current()
	dir := usr.HomeDir
	configFileApi := dir + "/config/secrets/api_" + strings.ToLower(exchange)
	err := gonfig.GetConf(configFileApi, &configApi)
	return &configApi, err
}
