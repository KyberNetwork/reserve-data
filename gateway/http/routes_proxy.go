package http

//Option define initialize behavior for server
type Option func(*Server) error

// WithCoreEndpoint set endpoint gateway for V3
func WithCoreEndpoint(coreEndpoint string) Option {
	return func(s *Server) error {
		coreProxyMW, err := newReverseProxyMW(coreEndpoint)
		if err != nil {
			return err
		}
		g := s.r.Group("/v3")

		g.GET("/prices-version", coreProxyMW)
		g.GET("/prices", coreProxyMW)
		g.GET("/prices/:base/:quote", coreProxyMW)
		g.GET("/getrates", coreProxyMW)
		g.GET("/get-all-rates", coreProxyMW)

		g.GET("/authdata-version", coreProxyMW)
		g.GET("/authdata", coreProxyMW)
		g.GET("/activities", coreProxyMW)
		g.GET("/immediate-pending-activities", coreProxyMW)

		g.POST("/cancelorder", coreProxyMW)
		g.POST("/cancel-all-orders", coreProxyMW)
		g.POST("/deposit", coreProxyMW)
		g.POST("/withdraw", coreProxyMW)
		g.POST("/trade", coreProxyMW)
		g.POST("/setrates", coreProxyMW)
		g.GET("/tradehistory", coreProxyMW)

		g.GET("/timeserver", coreProxyMW)

		g.GET("/gold-feed", coreProxyMW)
		g.GET("/btc-feed", coreProxyMW)
		g.GET("/usd-feed", coreProxyMW)
		g.POST("/set-feed-configuration", coreProxyMW)
		g.GET("/get-feed-configuration", coreProxyMW)

		g.GET("/addresses", coreProxyMW)

		return nil
	}
}

//WithSettingEndpoint set endpoint gateway for V3
func WithSettingEndpoint(settingEndpoint string) Option {
	return func(s *Server) error {
		settingProxyMW, err := newReverseProxyMW(settingEndpoint)
		if err != nil {
			return err
		}
		g := s.r.Group("/v3")

		g.GET("/asset/:id", settingProxyMW)
		g.GET("/asset", settingProxyMW)
		g.GET("/exchange/:id", settingProxyMW)
		g.GET("/exchange", settingProxyMW)
		g.GET("trading-pair/:id", settingProxyMW)
		g.GET("/stable-token-params", settingProxyMW)

		g.GET("/setting-change-main", settingProxyMW)
		g.GET("setting-change-main/:id", settingProxyMW)
		g.POST("/setting-change-main", settingProxyMW)
		g.PUT("/setting-change-main/:id", settingProxyMW)
		g.DELETE("/setting-change-main/:id", settingProxyMW)

		g.GET("/setting-change-target", settingProxyMW)
		g.GET("setting-change-target/:id", settingProxyMW)
		g.POST("/setting-change-target", settingProxyMW)
		g.PUT("/setting-change-target/:id", settingProxyMW)
		g.DELETE("/setting-change-target/:id", settingProxyMW)

		g.GET("/setting-change-rbquadratic", settingProxyMW)
		g.GET("setting-change-rbquadratic/:id", settingProxyMW)
		g.POST("/setting-change-rbquadratic", settingProxyMW)
		g.PUT("/setting-change-rbquadratic/:id", settingProxyMW)
		g.DELETE("/setting-change-rbquadratic/:id", settingProxyMW)

		g.GET("/setting-change-pwis", settingProxyMW)
		g.GET("setting-change-pwis/:id", settingProxyMW)
		g.POST("/setting-change-pwis", settingProxyMW)
		g.PUT("/setting-change-pwis/:id", settingProxyMW)
		g.DELETE("/setting-change-pwis/:id", settingProxyMW)

		g.GET("/setting-change-stable", settingProxyMW)
		g.GET("setting-change-stable/:id", settingProxyMW)
		g.POST("/setting-change-stable", settingProxyMW)
		g.PUT("/setting-change-stable/:id", settingProxyMW)
		g.DELETE("/setting-change-stable/:id", settingProxyMW)

		g.GET("/setting-change-update-exchange", settingProxyMW)
		g.GET("setting-change-update-exchange/:id", settingProxyMW)
		g.POST("/setting-change-update-exchange", settingProxyMW)
		g.PUT("/setting-change-update-exchange/:id", settingProxyMW)
		g.DELETE("/setting-change-update-exchange/:id", settingProxyMW)

		g.GET("/rebalance-status", settingProxyMW)
		g.POST("/hold-rebalance", settingProxyMW)
		g.POST("/enable-rebalance", settingProxyMW)

		g.GET("/set-rate-status", settingProxyMW)
		g.POST("/hold-set-rate", settingProxyMW)
		g.POST("/enable-set-rate", settingProxyMW)

		g.GET("/price-factor", settingProxyMW)
		g.POST("/price-factor", settingProxyMW)

		return nil
	}
}
