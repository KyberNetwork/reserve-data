package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/spf13/cobra"

	"github.com/KyberNetwork/reserve-data/settings"
	"github.com/KyberNetwork/reserve-data/settings/storage"
)

var (
	settingDB string
	exchanges string
)

func listNotInternalToken(cmd *cobra.Command, args []string) {
	if ii, err := os.Stat(settingDB); err != nil || ii.IsDir() {
		panic(fmt.Sprintf("db file %s not exists", settingDB))
	}
	bs, err := storage.NewBoltSettingStorage(settingDB)
	if err != nil {
		panic(err)
	}
	tokens, err := bs.GetAllTokens()
	if err != nil {
		log.Panicln(err)
	}
	fmt.Println("follow is list of all NOT internal token, can consider to be unlist")
	fmt.Printf("ID\tName\tAddress\tIsActive\n")
	for _, tk := range tokens {
		if !tk.Internal {
			fmt.Printf("%s\t%s\t%s\t%v\n", tk.ID, tk.Name, tk.Address, tk.Active)
		}
	}
}

func removeTokenInfo(cmd *cobra.Command, args []string) {
	var err error
	var db *bolt.DB
	db, err = bolt.Open(settingDB, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = db.Close()
	}()
	var exs []settings.ExchangeName
	for _, e := range strings.Split(exchanges, ",") {
		exh, ok := settings.ExchangeTypeValues()[e]
		if ok {
			exs = append(exs, exh)
		} else {
			panic(fmt.Sprintf("exchange not supported %s", e))
		}
	}
	err = db.Update(func(tx *bolt.Tx) error {
		return storage.RemoveTokensFromExchanges(tx, args, exs)
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("unlist token %v successfully\n", args)
}

func main() {

	rootCmd := &cobra.Command{
		Use:     "unlisted-token-fix",
		Example: "unlisted-token-fix",
		Run:     listNotInternalToken,
	}
	rootCmd.PersistentFlags().StringVar(&settingDB, "setting-db", "dev_setting.db", "setting db file path")
	unlistToken := &cobra.Command{
		Use:     "unlist [token]",
		Example: "unlist --exchanges binance,huobi,stable_exchange knc",
		Run:     removeTokenInfo,
	}
	unlistToken.LocalFlags().StringVar(&exchanges, "exchanges", "binance,huobi,stable_exchange", "a comma separate list of exchange to remove token info from")
	rootCmd.AddCommand(unlistToken)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
