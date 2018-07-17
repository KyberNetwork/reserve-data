package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/settings"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

func removeTokenFromList(tokens []common.Token, t common.Token) ([]common.Token, error) {
	if len(tokens) == 0 {
		return tokens, errors.New("Internal Token list is empty")
	}
	for i, token := range tokens {
		if token.ID == t.ID {
			tokens[len(tokens)-1], tokens[i] = tokens[i], tokens[len(tokens)-1]
			return tokens[:len(tokens)-1], nil
		}
	}
	return tokens, fmt.Errorf("The deactivating token %s is not in current internal token list", t.ID)
}

func (self *HTTPServer) reloadTokenIndices(newToken common.Token, active bool) error {
	tokens, err := self.setting.GetInternalTokens()
	if err != nil {
		return err
	}
	if active {
		tokens = append(tokens, newToken)
	} else {
		if tokens, err = removeTokenFromList(tokens, newToken); err != nil {
			return err
		}
	}
	if err = self.blockchain.LoadAndSetTokenIndices(common.GetTokenAddressesList(tokens)); err != nil {
		return err
	}
	return nil
}

func (self *HTTPServer) updateInternalTokensIndices(tokenListings map[string]common.TokenListing) error {
	tokens, err := self.setting.GetInternalTokens()
	if err != nil {
		return err
	}
	for _, tokenListing := range tokenListings {
		token := tokenListing.Token
		if token.Internal {
			tokens = append(tokens, token)
		}
	}
	if err = self.blockchain.LoadAndSetTokenIndices(common.GetTokenAddressesList(tokens)); err != nil {
		return err
	}
	return nil
}

// ensureRunningExchange makes sure that the exchange input is avaialbe in current deployment
func (self *HTTPServer) ensureRunningExchange(ex string) (settings.ExchangeName, error) {
	exName, ok := settings.ExchangeTypeValues()[ex]
	if !ok {
		return exName, fmt.Errorf("Exchange %s is not in current deployment", ex)
	}
	return exName, nil
}

// getExchangeSetting will query the current exchange setting with key ExName.
// return a struct contain all
func (self *HTTPServer) getExchangeSetting(exName settings.ExchangeName) (*common.ExchangeSetting, error) {
	exFee, err := self.setting.GetFee(exName)
	if err != nil {
		return nil, err
	}
	exMinDep, err := self.setting.GetMinDeposit(exName)
	if err != nil {
		return nil, err
	}
	exInfos, err := self.setting.GetExchangeInfo(exName)
	if err != nil {
		return nil, err
	}
	depAddrs, err := self.setting.GetDepositAddresses(exName)
	if err != nil {
		return nil, err
	}
	return common.NewExchangeSetting(depAddrs, exMinDep, exFee, exInfos), nil
}

func (self *HTTPServer) prepareExchangeSetting(token common.Token, tokExSetts map[string]common.TokenExchangeSetting, preparedExchangeSetting map[settings.ExchangeName]*common.ExchangeSetting) error {
	for ex, tokExSett := range tokExSetts {
		exName, err := self.ensureRunningExchange(ex)
		if err != nil {
			return fmt.Errorf("Exchange %s is not in current deployment", ex)
		}
		comExSet, ok := preparedExchangeSetting[exName]
		//create a current ExchangeSetting from setting if it does not exist yet
		if !ok {
			comExSet, err = self.getExchangeSetting(exName)
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

// ListToken will pre-process the token request and put into pending token request
// It will not apply any change to DB if the request is not as dictated in documentation.
// Newer request will append if the tokenID is not avail in pending, and overwrite otherwise
func (self *HTTPServer) ListToken(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"data"}, []Permission{ConfigurePermission})
	if !ok {
		return
	}
	data := []byte(postForm.Get("data"))
	tokenListings := make(map[string]common.TokenListing)
	if err := json.Unmarshal(data, &tokenListings); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("cant not unmarshall token request (%s)", err.Error())))
		return
	}

	var (
		exInfos map[string]common.ExchangeInfo
		err     error
	)
	hasInternal := thereIsInternal(tokenListings)
	if hasInternal {
		if self.hasMetricPending() {
			httputil.ResponseFailure(c, httputil.WithReason("There is currently pending action on metrics. Clean it first"))
			return
		}
		// verify exchange status and exchange precision limit for each exchange
		exInfos, err = self.getInfosFromExchangeEndPoint(tokenListings)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
	}
	// prepare each TokenListing instance for individual token
	for tokenID, tokenlisting := range tokenListings {
		token := tokenlisting.Token
		// To list token, its active status must be true
		token.Active = true
		// if the token is internal, it must come with PWIEq, targetQty and QuadraticEquation and exchange setting
		if token.Internal {
			if uErr := self.ensureInternalSetting(tokenlisting); uErr != nil {
				httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Token %s is internal, required more setting (%s)", token.ID, uErr.Error())))
				return
			}

			for ex, tokExSett := range tokenlisting.Exchanges {
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
				tokenlisting.Exchanges[ex] = tokExSett
			}
		}
		tokenListings[tokenID] = tokenlisting
	}

	if err = self.setting.UpdatePendingTokenListings(tokenListings); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) GetPendingTokenListings(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := self.setting.GetPendingTokenListings()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
	return
}

func (self *HTTPServer) ConfirmTokenListing(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"data"}, []Permission{ConfigurePermission})
	if !ok {
		return
	}
	data := []byte(postForm.Get("data"))
	var tokenListings map[string]common.TokenListing
	if err := json.Unmarshal(data, &tokenListings); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("cant not unmarshall token request %s", err.Error())))
		return
	}
	var (
		pws    common.PWIEquationRequestV2
		tarQty common.TokenTargetQtyV2
		quadEq common.RebalanceQuadraticRequest
		err    error
	)
	hasInternal := thereIsInternal(tokenListings)
	//If there is internal token in the listing, query for related information.
	if hasInternal {
		pws, err = self.metric.GetPWIEquationV2()
		if err != nil {
			log.Printf("WARNING: There is no current PWS equation in database, creating new instance...")
			pws = make(common.PWIEquationRequestV2)
		}
		tarQty, err = self.metric.GetTargetQtyV2()
		if err != nil {
			log.Printf("WARNING: There is no current target quantity in database, creating new instance...")
			tarQty = make(common.TokenTargetQtyV2)
		}
		quadEq, err = self.metric.GetRebalanceQuadratic()
		if err != nil {
			log.Printf("WARNING: There is no current quadratic equation in database, creating new instance...")
			quadEq = make(common.RebalanceQuadraticRequest)
		}
		if self.hasMetricPending() {
			httputil.ResponseFailure(c, httputil.WithReason("There is currently pending action on metrics. Clean it first"))
			return
		}
	}
	pendingTLs, err := self.setting.GetPendingTokenListings()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Can not get pending token listing (%s)", err.Error())))
		return
	}

	//Prepare all the object to update core setting to newly listed token
	preparedExchangeSetting := make(map[settings.ExchangeName]*common.ExchangeSetting)
	preparedToken := []common.Token{}
	for tokenID, tokenListing := range tokenListings {
		token := tokenListing.Token
		token.LastActivationChange = common.GetTimepoint()
		preparedToken = append(preparedToken, token)
		if token.Internal {
			//set metrics data for the token

			pws[tokenID] = tokenListing.PWIEq
			tarQty[tokenID] = tokenListing.TargetQty
			quadEq[tokenID] = tokenListing.QuadraticEq
		}
		//check if the token is available in pending token listing and is deep equal to it.
		pendingTL, avail := pendingTLs[tokenID]
		if !avail {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Token %s is not available in pending token listing", tokenID)))
			return
		}
		if eq := reflect.DeepEqual(pendingTL, tokenListing); !eq {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Confirm and pending token listing request for token %s are not equal", tokenID)))
			return
		}
		if uErr := self.prepareExchangeSetting(token, tokenListing.Exchanges, preparedExchangeSetting); uErr != nil {
			httputil.ResponseFailure(c, httputil.WithError(uErr))
			return
		}
	}
	//reload token indices and apply metric changes if the token is Internal
	if hasInternal {
		if err = self.updateInternalTokensIndices(tokenListings); err != nil {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Can not update internal token indices (%s)", err.Error())))
			return
		}
		if err = self.metric.ConfirmTokenListingInfo(tarQty, pws, quadEq); err != nil {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Can not update metric data (%s)", err.Error())))
			return
		}
	}
	// Apply the change into setting database
	if err = self.setting.ApplyTokenWithExchangeSetting(preparedToken, preparedExchangeSetting); err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Can not apply token and exchange setting for token listing (%s). Metric data and token indices changes has to be manually revert", err.Error())))
		return
	}

	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) RejectTokenListing(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	listings, err := self.setting.GetPendingTokenListings()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("there is no pending token listing (%s)", err.Error())))
		return
	}
	// TODO: Handling odd case when setting bucket DB op successful but metric bucket DB op failed.
	if err := self.setting.RemovePendingTokenListings(); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if thereIsInternal(listings) {
		if err := self.metric.RemovePendingTokenListingInfo(); err != nil {
			httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Token listing Removal has been finished but can not remove pending metrics data associated with the token listing (%s).This removal should be handle manually before calling token listing again. ", err.Error())))
			return
		}
	}
	httputil.ResponseSuccess(c)
}

// getInfosFromExchangeEndPoint assembles a map of exchange to lists of PairIDs and
// query their exchange Info in one go
func (self *HTTPServer) getInfosFromExchangeEndPoint(tokenListings map[string]common.TokenListing) (map[string]common.ExchangeInfo, error) {
	const ETHID = "ETH"
	exTokenPairIDs := make(map[string]([]common.TokenPairID))
	result := make(map[string]common.ExchangeInfo)
	for tokenID, TokenListing := range tokenListings {
		for ex, exSetting := range TokenListing.Exchanges {
			_, err := self.ensureRunningExchange(ex)
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
func (self *HTTPServer) hasMetricPending() bool {
	if _, err := self.metric.GetPendingPWIEquationV2(); err == nil {
		return true
	}
	if _, err := self.metric.GetPendingRebalanceQuadratic(); err == nil {
		return true
	}
	if _, err := self.metric.GetPendingTargetQtyV2(); err == nil {
		return true
	}
	return false
}

func thereIsInternal(tokenListings map[string]common.TokenListing) bool {
	for _, tokenListing := range tokenListings {
		if tokenListing.Token.Internal {
			return true
		}
	}
	return false
}

func (self *HTTPServer) ensureInternalSetting(tokenlisting common.TokenListing) error {
	token := tokenlisting.Token
	if uErr := self.blockchain.CheckTokenIndices(ethereum.HexToAddress(token.Address)); uErr != nil {
		return fmt.Errorf("cannot get token indice from smart contract (%s) ", uErr.Error())
	}
	if tokenlisting.Exchanges == nil {
		return errors.New("there is no exchange setting")
	}
	if tokenlisting.PWIEq == nil {
		return errors.New("there is no PWIS setting")
	}
	emptyTarget := common.TargetQtyV2{}
	if reflect.DeepEqual(emptyTarget, tokenlisting.TargetQty) {
		return errors.New("there is no target quantity setting or its values are all 0")
	}
	return nil
}

// UpdateToken update minor independent details about a token
// It provides the most simple way to modify  token's information without affecting other component
// To list/delist token, use ListToken/ DelistToken API instead.
func (self *HTTPServer) UpdateToken(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"data"}, []Permission{RebalancePermission, ConfigurePermission})
	if !ok {
		return
	}
	data := []byte(postForm.Get("data"))
	var token common.Token
	if err := json.Unmarshal(data, &token); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if _, err := self.setting.GetTokenByID(token.ID); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	//reload token indices if the token is Internal
	if token.Internal {
		if err := self.reloadTokenIndices(token, token.Internal); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
	}
	currTok, err := self.setting.GetTokenByID(token.ID)
	if err != nil && err != settings.ErrTokenNotFound {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	if err == settings.ErrTokenNotFound || currTok.Active != token.Active {
		token.LastActivationChange = common.GetTimepoint()
	} else {
		token.LastActivationChange = currTok.LastActivationChange
	}

	if err := self.setting.UpdateToken(token); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) TokenSettings(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	data, err := self.setting.GetAllTokens()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(data))
}

func (self *HTTPServer) UpdateAddress(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"name", "address"}, []Permission{RebalancePermission, ConfigurePermission})
	if !ok {
		return
	}
	addrStr := postForm.Get("address")
	name := postForm.Get("name")
	addressName, ok := settings.AddressNameValues()[name]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("invalid address name: %s", name)))
		return
	}
	addr := ethereum.HexToAddress(addrStr)
	if err := self.setting.UpdateAddress(addressName, addr); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) AddAddressToSet(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"name", "address"}, []Permission{RebalancePermission, ConfigurePermission})
	if !ok {
		return
	}
	addrStr := postForm.Get("address")
	addr := ethereum.HexToAddress(addrStr)
	setName := postForm.Get("name")
	addrSetName, ok := settings.AddressSetNameValues()[setName]
	if !ok {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("invalid address set name: %s", setName)))
		return
	}
	if err := self.setting.AddAddressToSet(addrSetName, addr); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) UpdateExchangeFee(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"name", "data"}, []Permission{RebalancePermission, ConfigurePermission})
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
	var exFee common.ExchangeFees
	if err := json.Unmarshal(data, &exFee); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if err := self.setting.UpdateFee(exName, exFee); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) UpdateExchangeMinDeposit(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"name", "data"}, []Permission{RebalancePermission, ConfigurePermission})
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
	var exMinDeposit common.ExchangesMinDeposit
	if err := json.Unmarshal(data, &exMinDeposit); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if err := self.setting.UpdateMinDeposit(exName, exMinDeposit); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) UpdateDepositAddress(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"name", "data"}, []Permission{RebalancePermission, ConfigurePermission})
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
	if err := self.setting.UpdateDepositAddress(exName, exDepositAddress); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) UpdateExchangeInfo(c *gin.Context) {
	postForm, ok := self.Authenticated(c, []string{"name", "data"}, []Permission{RebalancePermission, ConfigurePermission})
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
	var exInfo common.ExchangeInfo
	if err := json.Unmarshal(data, &exInfo); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if err := self.setting.UpdateExchangeInfo(exName, exInfo); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (self *HTTPServer) GetAllSetting(c *gin.Context) {
	_, ok := self.Authenticated(c, []string{}, []Permission{RebalancePermission, ConfigurePermission, ReadOnlyPermission, ConfirmConfPermission})
	if !ok {
		return
	}
	tokenSettings, err := self.setting.GetAllTokens()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	addressSettings, err := self.setting.GetAllAddresses()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	exchangeSettings := make(map[string]*common.ExchangeSetting)
	for exID := range common.SupportedExchanges {
		exName, vErr := self.ensureRunningExchange(string(exID))
		if vErr != nil {
			httputil.ResponseFailure(c, httputil.WithError(vErr))
			return
		}
		exSett, vErr := self.getExchangeSetting(exName)
		if vErr != nil {
			httputil.ResponseFailure(c, httputil.WithError(vErr))
			return
		}
		exchangeSettings[string(exID)] = exSett
	}

	allSetting := common.NewAllSettings(addressSettings, tokenSettings, exchangeSettings)
	httputil.ResponseSuccess(c, httputil.WithData(allSetting))
}
