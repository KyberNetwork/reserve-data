package marketdata

import (
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"

	marketdatacli "github.com/KyberNetwork/reserve-data/lib/market-data"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
	"github.com/pkg/errors"
)

type MarketData struct {
	l      *zap.SugaredLogger
	client *marketdatacli.Client
	s      storage.Interface
}

func New(client *marketdatacli.Client, s storage.Interface) *MarketData {
	return &MarketData{
		l:      zap.S(),
		s:      s,
		client: client,
	}
}

func dataForMarketDataByExchange(exchangeID rtypes.ExchangeID, base, quote string) (string, string, string, error) {
	var (
		lowerBase  = strings.ToLower(base)
		lowerQuote = strings.ToLower(quote)
	)
	publicSymbol := fmt.Sprintf("%s-%s", lowerBase, lowerQuote)
	switch exchangeID {
	case rtypes.Binance, rtypes.Binance2:
		return rtypes.Binance.String(), fmt.Sprintf("%s%s", lowerBase, lowerQuote), publicSymbol, nil
	case rtypes.Huobi:
		return rtypes.Huobi.String(), fmt.Sprintf("%s%s", lowerBase, lowerQuote), publicSymbol, nil
	default:
		return "", "", "", fmt.Errorf("%s exchange is not supported", exchangeID)
	}
}

func (md *MarketData) TryToAddFeed(data *common.AdditionalDataReturn) error {
	for _, tpID := range data.AddedTradingPairs {
		tradingPair, err := md.s.GetTradingPair(tpID, false)
		if err != nil {
			md.l.Errorw("cannot get trading pair", "id", tpID)
			return err
		}
		exchange, sourceSymbol, publicSymbol, err := dataForMarketDataByExchange(tradingPair.ExchangeID, tradingPair.BaseSymbol, tradingPair.QuoteSymbol)
		if err != nil {
			return err
		}
		if err := md.client.AddFeed(exchange, sourceSymbol, publicSymbol, strconv.FormatInt(int64(tpID), 10)); err != nil {
			md.l.Errorw("cannot add feed to market data", "err", err, "exchange", exchange, "source symbol", sourceSymbol)
			continue
		}
	}
	return nil
}

func (md *MarketData) ValidateSymbol(exchangeID rtypes.ExchangeID, baseSymbol, quoteSymbol string) error {
	exchange, symbol, _, err := dataForMarketDataByExchange(exchangeID, baseSymbol, quoteSymbol)
	if err != nil {
		return errors.Wrap(err, "cannot create params for market data client")
	}
	isValidSymbol, err := md.client.IsValidSymbol(exchange, symbol)
	if err != nil {
		return errors.Wrapf(err, "failed to verify pair %s on %s", symbol, exchange)
	}
	if !isValidSymbol {
		return errors.New(fmt.Sprintf("pair %s does not exists on %s", symbol, exchange))
	}
	return nil
}
