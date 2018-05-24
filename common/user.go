package common

type User struct {
}

type UserCap struct {
	UserID     string
	Category   string
	DailyLimit float64
	TxLimit    float64
	Type       string
}

func NonKycedCap() *UserCap {
	return &UserCap{
		Category:   "",
		DailyLimit: 15000.0,
		TxLimit:    3000.0,
		Type:       "non kyced",
	}
}

func KycedCap() *UserCap {
	return &UserCap{
		Category:   "",
		DailyLimit: 200000.0,
		TxLimit:    6000.0,
		Type:       "kyced",
	}
}
