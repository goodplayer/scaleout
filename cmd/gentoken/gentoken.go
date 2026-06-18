package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/meidoworks/nekoq-component/configure/permissions"
	"github.com/meidoworks/nekoq-component/configure/secret/api"
	"github.com/meidoworks/nekoq-component/configure/secret/impl"
	"github.com/meidoworks/nekoq-component/configure/secret/tools"
	"github.com/pelletier/go-toml/v2"

	"github.com/goodplayer/scaleout/consts"
)

type GenTokenConfig struct {
	PostgresConnectionString string `toml:"postgres-connection-string"`
}

var (
	configFilePath string
)

func init() {
	flag.StringVar(&configFilePath, "config", "gentoken.toml", "config file path")
	flag.Parse()
}

func main() {
	fn := func() []byte {
		f, err := os.Open(configFilePath)
		if err != nil {
			panic(err)
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)
		data, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}
		return data
	}
	cfg := new(GenTokenConfig)
	err := toml.Unmarshal(fn(), cfg)
	if err != nil {
		panic(err)
	}

	pgConnStr := cfg.PostgresConnectionString
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

	jwtTool := tools.NewJwtTool(keyStorage)
	jwtData := api.JwtData{}
	jwtTool.SetupPermissions(jwtData, tools.PermissionResourceList{}.
		Add(permissions.SecretJwtAdmin, permissions.SecretCertAdmin, permissions.SecretKeyAdmin))
	addon := tools.NewAddonTool(keyStorage)
	token, err := addon.SignJwtToken(consts.RootJwtTokenKey, api.JwtAlgHS512, tools.JwtClaims{}.FromJwtData(jwtData))
	if err != nil {
		panic(err)
	}
	fmt.Println("token:", token)
}

func checkNonEmptyString(str string) {
	if len(str) <= 0 {
		panic(errors.New("empty string"))
	}
}
