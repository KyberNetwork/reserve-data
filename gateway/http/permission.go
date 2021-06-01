package http

import (
	"fmt"

	"github.com/KyberNetwork/reserve-data/gateway/permission"
	"github.com/casbin/casbin"
	"github.com/gin-gonic/gin"
	scas "github.com/qiangmzsx/string-adapter"

	"github.com/KyberNetwork/httpsign-utils/authenticator"
)

var (
	readRole      = "read"
	manageRole    = "manage"
	rebalanceRole = "rebalance"
	adminRole     = "admin"
)

func addRoleKey(key, role string) string {
	return fmt.Sprintf(`
g, %s, %s`, key, role)
}

func addReadRolePolicy() string {
	return fmt.Sprintf(`
p, %[1]s, /*, GET`, readRole)
}

func addManageRolePolicy() string {
	return fmt.Sprintf(`
p, %[1]s, /*, GET
p, %[1]s, /v3/setting-change-update-exchange, POST
p, %[1]s, /v3/setting-change-target, POST
p, %[1]s, /v3/setting-change-pwis, POST
p, %[1]s, /v3/setting-change-rbquadratic, POST
p, %[1]s, /v3/setting-change-main, POST
p, %[1]s, /v3/setting-change-tpair, POST
p, %[1]s, /v3/setting-change-stable, POST
p, %[1]s, /v3/setting-change-feed-configuration, POST
p, %[1]s, /v3/setting-change-update-exchange/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-target/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-pwis/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-rbquadratic/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-main/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-tpair/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-stable/:id, (PUT)|(DELETE)
p, %[1]s, /v3/setting-change-feed-configuration/:id, (PUT)|(DELETE)
p, %[1]s, /v3/disapprove-setting-change/:id, DELETE
p, %[1]s, /v3/schedule-job/:id, DELETE`, manageRole)
}

func addAdminRolePolicy() string {
	return fmt.Sprintf(`
p, %[1]s, /*, GET
p, %[1]s, /v3/hold-rebalance, POST
p, %[1]s, /v3/enable-rebalance, POST
p, %[1]s, /v3/hold-set-rate, POST
p, %[1]s, /v3/gas-threshold, POST
p, %[1]s, /v3/set-exchange-enabled/:id, PUT
p, %[1]s, /v3/enable-set-rate, POST
p, %[1]s, /v3/rate-trigger-period, POST
p, %[1]s, /v3/gas-source, POST
p, %[1]s, /v3/update-feed-status/:name, PUT
p, %[1]s, /v3/schedule-job, POST`, adminRole)
}

func addRebalanceRolePolicy() string {
	return fmt.Sprintf(`
p, %[1]s, /*, GET
p, %[1]s, /v3/price-factor, POST
p, %[1]s, /v3/cancel-orders, POST
p, %[1]s, /v3/cancel-all-orders, POST
p, %[1]s, /v3/deposit, POST
p, %[1]s, /v3/withdraw, POST
p, %[1]s, /v3/trade, POST
p, %[1]s, /v3/cancel-setrates, POST
p, %[1]s, /transfer-self, POST
p, %[1]s, /cex-transfer, POST
p, %[1]s, /v3/setrates, POST`, rebalanceRole)
}

// NewPermissioner creates a gin Handle Func to controll permission
// currently there is only 2 permission for POST/GET requests
func NewPermissioner(readKeys, manageKeys, rebalanceKeys, adminRoles []authenticator.KeyPair) (gin.HandlerFunc, error) {
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
	pol += addReadRolePolicy() + addManageRolePolicy() + addRebalanceRolePolicy() + addAdminRolePolicy()
	for _, key := range readKeys {
		pol += addRoleKey(key.AccessKeyID, readRole)
	}
	for _, key := range manageKeys {
		pol += addRoleKey(key.AccessKeyID, manageRole)
	}
	for _, key := range rebalanceKeys {
		pol += addRoleKey(key.AccessKeyID, rebalanceRole)
	}
	for _, key := range adminRoles {
		pol += addRoleKey(key.AccessKeyID, adminRole)
	}
	sa := scas.NewAdapter(pol)
	e := casbin.NewEnforcer(casbin.NewModel(conf), sa)
	if err := e.LoadPolicy(); err != nil {
		return nil, err
	}
	p := permission.NewPermissioner(e)
	return p, nil
}
