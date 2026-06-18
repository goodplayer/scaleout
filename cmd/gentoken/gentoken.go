package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/meidoworks/nekoq-component/configure/configclient"
	"github.com/meidoworks/nekoq-component/configure/secret/api"
	"github.com/meidoworks/nekoq-component/configure/secret/impl"
	"github.com/meidoworks/nekoq-component/configure/secret/tools"

	"github.com/goodplayer/scaleout/consts"
)

func main() {
	cfgclient := configclient.NewEnvClient()
	pgConnStr := getEnvString(cfgclient, "POSTGRES_CONNECTION_STRING")
	checkNonEmptyString(pgConnStr)
	fmt.Println("debug pgConnStr:", pgConnStr)

	up, err := impl.NewLocalFileUnsealProvider(os.DirFS("."), map[int64]struct {
		KeyFilePath     string
		KeyFilePassword string
	}{
		1: {
			"bootstrap.key", "changeit",
		},
	})
	if err != nil {
		panic(err)
	}
	keyStorage, err := impl.NewPostgresKeyStorage(pgConnStr)
	if err != nil {
		panic(err)
	}
	if err := keyStorage.Startup(); err != nil {
		panic(err)
	}
	if err := keyStorage.SetupUnsealProviderAndWait(up); err != nil {
		panic(err)
	}
	fmt.Println("unseal success!")

	addon := tools.NewAddonTool(keyStorage)
	token, err := addon.SignJwtToken(consts.RootJwtTokenKey, api.JwtAlgHS512, tools.JwtClaims{
		"Hello": "World",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("token:", token)
}

func getEnvString(c *configclient.EnvClient, key string) string {
	val, err := c.GetString(key)
	if err != nil {
		panic(err)
	}
	return val
}

func checkNonEmptyString(str string) {
	if len(str) <= 0 {
		panic(errors.New("empty string"))
	}
}
