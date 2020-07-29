package http

import (
	"fmt"

	"github.com/KyberNetwork/reserve-data/gateway/permission"
	"github.com/casbin/casbin"
	"github.com/gin-gonic/gin"
	scas "github.com/qiangmzsx/string-adapter"

	"github.com/KyberNetwork/httpsign-utils/authenticator"
)

func addKeyReadPolicy(key string) string {
	return fmt.Sprintf(`
p, %[1]s, /*, GET`, key)
}

func addKeyWritePolicy(key string) string {
	return fmt.Sprintf(`
p, %[1]s, /*, GET
p, %[1]s, /v3/setting-change-update-exchange, POST
p, %[1]s, /v3/setting-change-target, POST
p, %[1]s, /v3/setting-change-pwis, POST
p, %[1]s, /v3/setting-change-rbquadratic, POST
p, %[1]s, /v3/setting-change-main, POST
p, %[1]s, /v3/setting-change-stable, POST
p, %[1]s, /v3/setting-change-feed-configuration, POST
p, %[1]s, /v3/update-feed-status/:name, PUT
p, %[1]s, /v3/rate-trigger-period-length, POST`, key)
}

func addKeyConfirmPolicy(key string) string {
	return fmt.Sprintf(`
p, %[1]s, /*, GET
p, %[1]s, /v3/setting-change-update-exchange/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-target/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-pwis/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-rbquadratic/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-main/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-stable/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-feed-configuration/:id, (PUT)|(DELETE)
p, %[1]s, /v3/hold-rebalance, POST
p, %[1]s, /v3/enable-rebalance, POST
p, %[1]s, /v3/hold-set-rate, POST
p, %[1]s, /v3/set-exchange-enabled/:id, PUT
p, %[1]s, /v3/enable-set-rate, POST`, key)
}

func addKeyRebalancePolicy(key string) string {
	return fmt.Sprintf(`
p, %[1]s, /*, GET
p, %[1]s, /v3/price-factor, POST
p, %[1]s, /v3/cancel-orders, POST
p, %[1]s, /v3/cancel-all-orders, POST
p, %[1]s, /v3/deposit, POST
p, %[1]s, /v3/withdraw, POST
p, %[1]s, /v3/trade, POST
p, %[1]s, /v3/setrates, POST`, key)
}

//NewPermissioner creates a gin Handle Func to controll permission
//currently there is only 2 permission for POST/GET requests
func NewPermissioner(readKeys, writeKeys, confirmKeys, rebalanceKeys []authenticator.KeyPair) (gin.HandlerFunc, error) {
	const (
		conf = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _ , _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub)  && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
`
	)
	var pol string
	for _, key := range readKeys {
		pol += addKeyReadPolicy(key.AccessKeyID)
	}
	for _, key := range writeKeys {
		pol += addKeyWritePolicy(key.AccessKeyID)
	}
	for _, key := range confirmKeys {
		pol += addKeyConfirmPolicy(key.AccessKeyID)
	}
	for _, key := range rebalanceKeys {
		pol += addKeyRebalancePolicy(key.AccessKeyID)
	}

	sa := scas.NewAdapter(pol)
	e := casbin.NewEnforcer(casbin.NewModel(conf), sa)
	if err := e.LoadPolicy(); err != nil {
		return nil, err
	}

	p := permission.NewPermissioner(e)
	return p, nil
}
