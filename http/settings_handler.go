package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/settings"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

const (
	pricingOPAddressName      = "pricing_operator"
	depositOPAddressName      = "deposit_operator"
	intermediateOPAddressName = "intermediate_operator"
	validAddressLength        = 42
)

func (hs *HTTPServer) updateInternalTokensIndices(tokenUpdates map[string]common.TokenUpdate) error {
	tokens, err := hs.setting.GetInternalTokens()
	if err != nil {
		return err
	}
	for _, tokenUpdate := range tokenUpdates {
		token := tokenUpdate.Token
		if token.Internal {
			tokens = append(tokens, token)
		}
	}
	if err = hs.blockchain.LoadAndSetTokenIndices(common.GetTokenAddressesList(tokens)); err != nil {
		return err
	}
	return nil
}

// ensureRunningExchange makes sure that the exchange input is avaialbe in current deployment
func (hs *HTTPServer) ensureRunningExchange(ex string) (settings.ExchangeName, error) {
	exName, ok := settings.ExchangeTypeValues()[ex]
	if !ok {
		return exName, fmt.Errorf("Exchange %s is not in current deployment", ex)
	}
	return exName, nil
}

// getExchangeSetting will query the current exchange setting with key ExName.
// return a struct contain all
func (hs *HTTPServer) getExchangeSetting(exName settings.ExchangeName) (*common.ExchangeSetting, error) {
	exFee, err := hs.setting.GetFee(exName)
	if err != nil {
		if err != settings.ErrExchangeRecordNotFound {
			return nil, err
		}
		log.Printf("the current exchange fee for %s hasn't existed yet.", exName.String())
		fundingFee := common.NewFundingFee(make(map[string]float64), make(map[string]float64))
		exFee = common.NewExchangeFee(make(common.TradingFee), fundingFee)
	}
	exMinDep, err := hs.setting.GetMinDeposit(exName)
	if err != nil {
		if err != settings.ErrExchangeRecordNotFound {
			return nil, err
		}
		log.Printf("the current exchange MinDeposit for %s hasn't existed yet.", exName.String())
		exMinDep = make(common.ExchangesMinDeposit)
	}
	exInfos, err := hs.setting.GetExchangeInfo(exName)
	if err != nil {
		if err != settings.ErrExchangeRecordNotFound {
			return nil, err
		}
		log.Printf("the current exchange Info for %s hasn't existed yet.", exName.String())
		exInfos = make(common.ExchangeInfo)
	}
	depAddrs, err := hs.setting.GetDepositAddresses(exName)
	if err != nil {
		if err != settings.ErrExchangeRecordNotFound {
			return nil, err
		}
		log.Printf("the current exchange deposit addresses  for %s hasn't existed yet.", exName.String())
		depAddrs = make(common.ExchangeAddresses)
	}
	return common.NewExchangeSetting(depAddrs, exMinDep, exFee, exInfos), nil
}

func (hs *HTTPServer) prepareExchangeSetting(token common.Token, tokExSetts map[string]common.TokenExchangeSetting, preparedExchangeSetting map[settings.ExchangeName]*common.ExchangeSetting) error {
	for ex, tokExSett := range tokExSetts {
		exName, err := hs.ensureRunningExchange(ex)
		if err != nil {
			return fmt.Errorf("Exchange %s is not in current deployment", ex)
		}
		comExSet, ok := preparedExchangeSetting[exName]
		//create a current ExchangeSetting from setting if it does not exist yet
		if !ok {
			comExSet, err = hs.getExchangeSetting(exName)
			if err != nil {
				return err
			}
		}
		//update exchange Deposite Address for ExchangeSetting
		comExSet.DepositAddress[token.ID] = ethereum.HexToAddress(tokExSett.DepositAddress)

		//update Exchange Info for ExchangeSetting
		for pairID, epl := range tokExSett.Info {
			comExSet.Info[pairID] = epl
		}

		//Update Exchange Fee for ExchangeSetting
		comExSet.Fee.Funding.Deposit[token.ID] = tokExSett.Fee.Deposit
		comExSet.Fee.Funding.Withdraw[token.ID] = tokExSett.Fee.Withdraw

		//Update Exchange Min deposit for ExchangeSetting
		comExSet.MinDeposit[token.ID] = tokExSett.MinDeposit

		preparedExchangeSetting[exName] = comExSet
	}
	return nil
}

// SetTokenUpdate will pre-process the token request and put into pending token request
// It will not apply any change to DB if the request is not as dictated in documentation.
// Newer request will append if the tokenID is not avail in pending, and overwrite otherwise
func (hs *HTTPServer) SetTokenUpdate(c *gin.Context) {
	postForm, ok := hs.Authenticated(c, []string{"data"}, []Permission{ConfigurePermission})
	if !ok {
		return
	}
	data := []byte(postForm.Get("data"))
	tokenUpdates := make(map[string]common.TokenUpdate)
	if err := json.Unmarshal(data, &tokenUpdates); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("cant not unmarshall token request (%s)", err.Error())))
		return
	}

	var (
		exInfos map[string]common.ExchangeInfo
		err     error
	)
	hasInternal := thereIsInternal(tokenUpdates)
	if hasInternal {
		if hs.hasMetricPending() {
			httputil.ResponseFailure(c, httputil.WithReason("There is currently pending action on metrics. Clean it first"))
			return
		}
		// verify exchange status and exchange precision limit for each exchange
		exInfos, err = hs.getInfosFromExchangeEndPoint(tokenUpdates)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
	}
	// prepare each tokenUpdate instance for individual token
	for tokenID, tokenUpdate := range tokenUpdates {
		token := tokenUpdate.Token
		token.ID = tokenID
		if len(token.Address) != validAddressLength {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Token %s's address is invalid length. Token field in token request might be empty ", tokenID)))
			return
		}
		if len(token.Name) == 0 {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Token %s's name is empty. Token field in token request might be empty ", tokenID)))
			return
		}
		// if the token is internal, it must come with PWIEq, targetQty and QuadraticEquation and exchange setting
		if token.Internal {
			if uErr := hs.ensureInternalSetting(tokenUpdate); uErr != nil {
				httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Token %s is internal, required more setting (%s)", token.ID, uErr.Error())))
				return
			}

			for ex, tokExSett := range tokenUpdate.Exchanges {
				//query exchangeprecisionlimit from exchange for the pair token-ETH
				pairID := common.NewTokenPairID(token.ID, "ETH")
				// If the pair is not in current token listing request, get its result from exchange
				_, ok1 := tokExSett.Info[pairID]
				if !ok1 {
					if tokExSett.Info == nil {
						tokExSett.Info = make(common.ExchangeInfo)
					}
					exchangePrecisionLimit, ok2 := exInfos[ex].GetData()[pairID]
					if !ok2 {
						httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Pair ID %s on exchange %s couldn't be queried for exchange presicion limit. This info must be manually entered", pairID, ex)))
						return
					}
					tokExSett.Info[pairID] = exchangePrecisionLimit
				}
				tokenUpdate.Exchanges[ex] = tokExSett
			}
		}
		tokenUpdate.Token = token
		tokenUpdates[tokenID] = tokenUpdate
	}

	if err = hs.setting.UpdatePendingTokenUpdates(tokenUpdates); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (hs *HTTPServer) GetPendingTokenUpdates(c *gin.Context) {
	_, ok := hs.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := hs.setting.GetPendingTokenUpdates()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
	return
}

func (hs *HTTPServer) ConfirmTokenUpdate(c *gin.Context) {
	postForm, ok := hs.Authenticated(c, []string{"data"}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	data := []byte(postForm.Get("data"))
	//no need to handle error here, if timestamp==0 the program will use UNIX timestamp instead
	timestamp, _ := strconv.ParseUint(postForm.Get("timestamp"), 10, 64)
	var tokenUpdates map[string]common.TokenUpdate
	if err := json.Unmarshal(data, &tokenUpdates); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("cant not unmarshall token request %s", err.Error())))
		return
	}
	var (
		pws    common.PWIEquationRequestV2
		tarQty common.TokenTargetQtyV2
		quadEq common.RebalanceQuadraticRequest
		err    error
	)
	hasInternal := thereIsInternal(tokenUpdates)
	//If there is internal token in the listing, query for related information.
	if hasInternal {
		pws, err = hs.metric.GetPWIEquationV2()
		if err != nil {
			log.Printf("WARNING: There is no current PWS equation in database, creating new instance...")
			pws = make(common.PWIEquationRequestV2)
		}
		tarQty, err = hs.metric.GetTargetQtyV2()
		if err != nil {
			log.Printf("WARNING: There is no current target quantity in database, creating new instance...")
			tarQty = make(common.TokenTargetQtyV2)
		}
		quadEq, err = hs.metric.GetRebalanceQuadratic()
		if err != nil {
			log.Printf("WARNING: There is no current quadratic equation in database, creating new instance...")
			quadEq = make(common.RebalanceQuadraticRequest)
		}
		if hs.hasMetricPending() {
			httputil.ResponseFailure(c, httputil.WithReason("There is currently pending action on metrics. Clean it first"))
			return
		}
	}
	pendingTLs, err := hs.setting.GetPendingTokenUpdates()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Can not get pending token listing (%s)", err.Error())))
		return
	}

	//Prepare all the object to update core setting to newly listed token
	preparedExchangeSetting := make(map[settings.ExchangeName]*common.ExchangeSetting)
	preparedToken := []common.Token{}
	for tokenID, tokenUpdate := range tokenUpdates {
		token := tokenUpdate.Token
		token.LastActivationChange = common.GetTimepoint()
		preparedToken = append(preparedToken, token)
		if token.Internal {
			//set metrics data for the token

			pws[tokenID] = tokenUpdate.PWIEq
			tarQty[tokenID] = tokenUpdate.TargetQty
			quadEq[tokenID] = tokenUpdate.QuadraticEq
		}
		//check if the token is available in pending token listing and is deep equal to it.
		pendingTL, avail := pendingTLs[tokenID]
		if !avail {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Token %s is not available in pending token listing", tokenID)))
			return
		}
		if eq := reflect.DeepEqual(pendingTL, tokenUpdate); !eq {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Confirm and pending token listing request for token %s are not equal", tokenID)))
			return
		}
		if uErr := hs.prepareExchangeSetting(token, tokenUpdate.Exchanges, preparedExchangeSetting); uErr != nil {
			httputil.ResponseFailure(c, httputil.WithError(uErr))
			return
		}
	}
	//reload token indices and apply metric changes if the token is Internal
	if hasInternal {
		if err = hs.updateInternalTokensIndices(tokenUpdates); err != nil {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Can not update internal token indices (%s)", err.Error())))
			return
		}
		if err = hs.metric.ConfirmTokenUpdateInfo(tarQty, pws, quadEq); err != nil {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Can not update metric data (%s)", err.Error())))
			return
		}
	}
	// Apply the change into setting database
	if err = hs.setting.ApplyTokenWithExchangeSetting(preparedToken, preparedExchangeSetting, timestamp); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Can not apply token and exchange setting for token listing (%s). Metric data and token indices changes has to be manually revert", err.Error())))
		return
	}

	httputil.ResponseSuccess(c)
}

func (hs *HTTPServer) RejectTokenUpdate(c *gin.Context) {
	_, ok := hs.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	listings, err := hs.setting.GetPendingTokenUpdates()
	if (err != nil) || len(listings) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("there is no pending token listing (%v)", err)))
		return
	}
	// TODO: Handling odd case when setting bucket DB op successful but metric bucket DB op failed.
	if err := hs.setting.RemovePendingTokenUpdates(); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

// getInfosFromExchangeEndPoint assembles a map of exchange to lists of PairIDs and
// query their exchange Info in one go
func (hs *HTTPServer) getInfosFromExchangeEndPoint(tokenUpdates map[string]common.TokenUpdate) (map[string]common.ExchangeInfo, error) {
	const ETHID = "ETH"
	exTokenPairIDs := make(map[string]([]common.TokenPairID))
	result := make(map[string]common.ExchangeInfo)
	for tokenID, tokenUpdate := range tokenUpdates {
		for ex, exSetting := range tokenUpdate.Exchanges {
			_, err := hs.ensureRunningExchange(ex)
			if err != nil {
				return result, err
			}
			info, ok := exTokenPairIDs[ex]
			if !ok {
				info = []common.TokenPairID{}
			}
			pairID := common.NewTokenPairID(tokenID, ETHID)
			//if the current exchangeSetting already got precision limit for this pair, skip it
			_, ok = exSetting.Info[pairID]
			if ok {
				continue
			}
			info = append(info, pairID)
			exTokenPairIDs[ex] = info
		}
	}
	for ex, tokenPairIDs := range exTokenPairIDs {
		exchange, err := common.GetExchange(ex)
		if err != nil {
			return result, err
		}
		liveInfo, err := exchange.GetLiveExchangeInfos(tokenPairIDs)
		if err != nil {
			return result, err
		}
		result[ex] = liveInfo
	}
	return result, nil
}

// hasPending return true if currently there is a pending request on Metric data
// This is to ensure that token listing operations does not conflict with metric operations
func (hs *HTTPServer) hasMetricPending() bool {
	if _, err := hs.metric.GetPendingPWIEquationV2(); err == nil {
		return true
	}
	if _, err := hs.metric.GetPendingRebalanceQuadratic(); err == nil {
		return true
	}
	if _, err := hs.metric.GetPendingTargetQtyV2(); err == nil {
		return true
	}
	return false
}

func thereIsInternal(tokenUpdates map[string]common.TokenUpdate) bool {
	for _, tokenUpdate := range tokenUpdates {
		if tokenUpdate.Token.Internal {
			return true
		}
	}
	return false
}

func (hs *HTTPServer) ensureInternalSetting(tokenUpdate common.TokenUpdate) error {
	token := tokenUpdate.Token
	if uErr := hs.blockchain.CheckTokenIndices(ethereum.HexToAddress(token.Address)); uErr != nil {
		return fmt.Errorf("cannot get token indice from smart contract (%s) ", uErr.Error())
	}
	if tokenUpdate.Exchanges == nil {
		return errors.New("there is no exchange setting")
	}
	if tokenUpdate.PWIEq == nil {
		return errors.New("there is no PWIS setting")
	}
	emptyTarget := common.TargetQtyV2{}
	if reflect.DeepEqual(emptyTarget, tokenUpdate.TargetQty) {
		return errors.New("there is no target quantity setting or its values are all 0")
	}
	return nil
}

func (hs *HTTPServer) TokenSettings(c *gin.Context) {
	_, ok := hs.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := hs.setting.GetAllTokens()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (hs *HTTPServer) UpdateExchangeFee(c *gin.Context) {
	postForm, ok := hs.Authenticated(c, []string{"name", "data"}, []Permission{RebalancePermission, ConfigurePermission})
	if !ok {
		return
	}
	name := postForm.Get("name")
	exName, ok := settings.ExchangeTypeValues()[name]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithError(fmt.Errorf("Exchange %s is not in current deployment", name)))
		return
	}
	data := []byte(postForm.Get("data"))
	//no need to handle error here, if timestamp==0 the program will use UNIX timestamp instead
	timestamp, _ := strconv.ParseUint(postForm.Get("timestamp"), 10, 64)
	var exFee common.ExchangeFees
	if err := json.Unmarshal(data, &exFee); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if exFee.Trading == nil || exFee.Funding.Deposit == nil || exFee.Funding.Withdraw == nil {
		httputil.ResponseFailure(c, httputil.WithReason("Data is in the wrong format: there is nil map inside the data"))
	}
	if err := hs.setting.UpdateFee(exName, exFee, timestamp); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (hs *HTTPServer) UpdateExchangeMinDeposit(c *gin.Context) {
	postForm, ok := hs.Authenticated(c, []string{"name", "data"}, []Permission{RebalancePermission, ConfigurePermission})
	if !ok {
		return
	}
	name := postForm.Get("name")
	exName, ok := settings.ExchangeTypeValues()[name]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Exchange %s is not in current deployment", name)))
		return
	}
	data := []byte(postForm.Get("data"))
	//no need to handle error here, if timestamp==0 the program will use UNIX timestamp instead
	timestamp, _ := strconv.ParseUint(postForm.Get("timestamp"), 10, 64)
	var exMinDeposit common.ExchangesMinDeposit
	if err := json.Unmarshal(data, &exMinDeposit); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if err := hs.setting.UpdateMinDeposit(exName, exMinDeposit, timestamp); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (hs *HTTPServer) UpdateDepositAddress(c *gin.Context) {
	postForm, ok := hs.Authenticated(c, []string{"name", "data"}, []Permission{RebalancePermission, ConfigurePermission})
	if !ok {
		return
	}
	name := postForm.Get("name")
	exName, ok := settings.ExchangeTypeValues()[name]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Exchange %s is not in current deployment", name)))
		return
	}
	data := []byte(postForm.Get("data"))
	//no need to handle error here, if timestamp==0 the program will use UNIX timestamp instead
	timestamp, _ := strconv.ParseUint(postForm.Get("timestamp"), 10, 64)
	var exDepositAddressStr map[string]string
	if err := json.Unmarshal(data, &exDepositAddressStr); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	exDepositAddress := make(common.ExchangeAddresses)
	for tokenID, addrStr := range exDepositAddressStr {
		log.Printf("addrstr is %s", addrStr)
		exDepositAddress[tokenID] = ethereum.HexToAddress(addrStr)
		log.Printf(exDepositAddress[tokenID].Hex())
	}
	if err := hs.setting.UpdateDepositAddress(exName, exDepositAddress, timestamp); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (hs *HTTPServer) UpdateExchangeInfo(c *gin.Context) {
	postForm, ok := hs.Authenticated(c, []string{"name", "data"}, []Permission{RebalancePermission, ConfigurePermission})
	if !ok {
		return
	}
	name := postForm.Get("name")
	exName, ok := settings.ExchangeTypeValues()[name]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Exchange %s is not in current deployment", name)))
		return
	}
	data := []byte(postForm.Get("data"))
	//no need to handle error here, if timestamp==0 the program will use UNIX timestamp instead
	timestamp, _ := strconv.ParseUint(postForm.Get("timestamp"), 10, 64)
	var exInfo common.ExchangeInfo
	if err := json.Unmarshal(data, &exInfo); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	// go through the input data and find if there is any empty Exchange precision limit
	tokenPairIDlist := []common.TokenPairID{}
	emptyEPL := common.ExchangePrecisionLimit{}
	for pairID, ei := range exInfo {
		if reflect.DeepEqual(ei, emptyEPL) {
			tokenPairIDlist = append(tokenPairIDlist, pairID)
		}
	}

	// If there is tokenPairID with empty exchange info, query and set it into the exchange info to be update
	if len(tokenPairIDlist) > 0 {
		exchange, err := common.GetExchange(name)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		liveInfo, err := exchange.GetLiveExchangeInfos(tokenPairIDlist)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("There is empty exchange info but can not get live exchange info fom exchange (%s) ", err.Error())))
		}
		for tokenPairID, epl := range liveInfo {
			exInfo[tokenPairID] = epl
		}
	}
	if err := hs.setting.UpdateExchangeInfo(exName, exInfo, timestamp); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (hs *HTTPServer) GetAllSetting(c *gin.Context) {
	_, ok := hs.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	addrReponse, err := hs.getAddressResponse()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	tokResponse, err := hs.getTokenResponse()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	exchangeResponse, err := hs.getExchangeResponse()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	timepoint := common.GetTimepoint()
	allSetting := common.NewAllSettings(addrReponse, tokResponse, exchangeResponse)
	httputil.ResponseSuccess(c, httputil.WithMultipleFields(gin.H{
		"timestamp": timepoint,
		"data":      allSetting,
	}))
}

func (hs *HTTPServer) getAddressResponse() (*common.AddressesResponse, error) {
	addressSettings, err := hs.setting.GetAllAddresses()
	if err != nil {
		return nil, err
	}
	addressSettings[pricingOPAddressName] = hs.blockchain.GetPricingOPAddress().Hex()
	addressSettings[depositOPAddressName] = hs.blockchain.GetDepositOPAddress().Hex()
	if _, ok := common.SupportedExchanges["huobi"]; ok {
		addressSettings[intermediateOPAddressName] = hs.blockchain.GetIntermediatorOPAddress().Hex()
	}
	if err != nil {
		return nil, err
	}
	addressResponse := common.NewAddressResponse(addressSettings)
	return addressResponse, nil
}

func (hs *HTTPServer) getTokenResponse() (*common.TokenResponse, error) {
	tokens, err := hs.setting.GetAllTokens()
	if err != nil {
		return nil, err
	}
	version, err := hs.setting.GetTokenVersion()
	if err != nil {
		return nil, err
	}
	tokenResponse := common.NewTokenResponse(tokens, version)
	return tokenResponse, nil
}

func (hs *HTTPServer) getExchangeResponse() (*common.ExchangeResponse, error) {
	exchangeSettings := make(map[string]*common.ExchangeSetting)
	for exID := range common.SupportedExchanges {
		exName, err := hs.ensureRunningExchange(string(exID))
		if err != nil {
			return nil, err
		}
		exSett, err := hs.getExchangeSetting(exName)
		if err != nil {
			return nil, err
		}
		exchangeSettings[string(exID)] = exSett
	}
	version, err := hs.setting.GetExchangeVersion()
	if err != nil {
		return nil, err
	}
	return common.NewExchangeResponse(exchangeSettings, version), nil
}

func (hs *HTTPServer) GetInternalTokens(c *gin.Context) {
	_, ok := hs.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	tokens, err := hs.setting.GetInternalTokens()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(tokens))
}

func (hs *HTTPServer) GetActiveTokens(c *gin.Context) {
	_, ok := hs.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	tokens, err := hs.setting.GetActiveTokens()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(tokens))
}

func (hs *HTTPServer) GetTokenByAddress(c *gin.Context) {
	_, ok := hs.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	addr := c.Query("address")
	if len(addr) != validAddressLength {
		httputil.ResponseFailure(c, httputil.WithReason("address is invalid length "))
		return
	}
	token, err := hs.setting.GetTokenByAddress(ethereum.HexToAddress(addr))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(token))
}

func (hs *HTTPServer) GetActiveTokenByID(c *gin.Context) {
	_, ok := hs.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	ID := c.Query("ID")
	token, err := hs.setting.GetActiveTokenByID((ID))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(token))
}

func (hs *HTTPServer) GetAddress(c *gin.Context) {
	_, ok := hs.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	name := c.Query("name")
	addressNames := settings.AddressNameValues()
	addrName, ok := addressNames[name]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("address name %s is not avail in this list of valid address name", name)))
		return
	}
	address, err := hs.setting.GetAddress(addrName)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(address))
}

func (hs *HTTPServer) GetAddresses(c *gin.Context) {
	_, ok := hs.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	name := c.Query("name")
	addressSetNames := settings.AddressSetNameValues()
	addrSetName, ok := addressSetNames[name]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("address set name %s is not avail in this list of valid address set name", name)))
		return
	}
	address, err := hs.setting.GetAddresses(addrSetName)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(address))
}

func (hs *HTTPServer) ReadyToServe(c *gin.Context) {
	_, ok := hs.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	httputil.ResponseSuccess(c)
}
