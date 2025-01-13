package main

import (
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/meidoworks/nekoq-component/configure/cfgimpl"
	"github.com/meidoworks/nekoq-component/configure/configapi"
	"github.com/meidoworks/nekoq-component/configure/configclient"
	"github.com/meidoworks/nekoq-component/configure/configserver"
	"github.com/meidoworks/nekoq-component/configure/secret"
	"github.com/meidoworks/nekoq-component/configure/secretapi"
	"github.com/meidoworks/nekoq-component/configure/secretimpl"

	"github.com/goodplayer/scaleout/consts"
	"github.com/goodplayer/scaleout/container"
	"github.com/goodplayer/scaleout/scaleout"
)

type SecretStorage interface {
	secretapi.KeyStorage
	secretapi.CertStorage
}

type configures struct {
}

func main() {
	cfgclient := configclient.NewEnvClient()
	debug, err := cfgclient.GetBool("DEBUG")
	if err != nil && !errors.Is(err, configclient.ErrNoSuchEnvironmentVariable) {
		panic(err)
	}
	pgConnStr := getEnvString(cfgclient, "POSTGRES_CONNECTION_STRING")
	checkNonEmptyString(pgConnStr)
	if debug {
		fmt.Println("debug pgConnStr:", pgConnStr)
	}

	// secret initialization
	keyStorage := initializeKeyStorage(pgConnStr)
	tlsCert, tlsCertKeyInfo, err := keyStorage.LoadCertByName(consts.RootClusterTLSCert, secretapi.CertLevelTypeCert)
	if err != nil {
		panic(err)
	}
	tlsPri := loadCertKey(keyStorage, tlsCertKeyInfo)

	con := container.NewContainer()

	// configure components
	con.AddLifecycle(&scaleout.Configure{
		PrepareFn: func(ctx context.Context) (context.Context, error) {
			dataPump := cfgimpl.NewDatabaseDataPump(pgConnStr)
			dataWriter := cfgimpl.NewDatabaseDataWriter(pgConnStr)
			srv := configserver.NewConfigureServer(configserver.ConfigureOptions{
				Addr: ":4001",
				TLSConfig: struct {
					Addr string
					Cert *x509.Certificate
					Key  crypto.PrivateKey
				}{
					Addr: ":4002",
					Cert: tlsCert,
					Key:  tlsPri,
				},
				WriteApi: struct {
					DataWriter configapi.DataWriter
					Addr       string
					TLSConfig  struct {
						Addr string
						Cert *x509.Certificate
						Key  crypto.PrivateKey
					}
				}{
					DataWriter: dataWriter,
					Addr:       ":4003",
					TLSConfig: struct {
						Addr string
						Cert *x509.Certificate
						Key  crypto.PrivateKey
					}{
						Addr: ":4004",
						Cert: tlsCert,
						Key:  tlsPri,
					},
				},
				DataPump: dataPump,
			})
			newCtx := context.WithValue(ctx, "server", srv)
			return newCtx, nil
		},
		OnStartFn: func(ctx context.Context) error {
			val := ctx.Value("server").(*configserver.ConfigureServer)
			return val.Startup()
		},
		OnStopFn: func(ctx context.Context) error {
			log.Println("shutting down configure server...")
			val := ctx.Value("server").(*configserver.ConfigureServer)
			return val.Shutdown()
		},
	})

	if err := con.Startup(); err != nil {
		panic(err)
	}
	log.Println("container started successfully!")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	log.Println("signal received:", sig)
	log.Println("shutting down...")
	if err := con.Shutdown(); err != nil {
		panic(err)
	}
	log.Println("shutdown complete.")
}

func loadCertKey(storage SecretStorage, info secretapi.CertKeyInfo) crypto.PrivateKey {
	// currently supporting level2 key as TLS cert key
	switch info.CertKeyLevel {
	case secretapi.CertKeyLevelUnseal:
		panic(errors.New("unseal certificate key level detected"))
	case secretapi.CertKeyLevelLevel1Ecdsa:
		panic(errors.New("level1 ecdsa certificate key level detected"))
	case secretapi.CertKeyLevelLevel1Rsa:
		panic(errors.New("level1 rsa certificate key level detected"))
	case secretapi.CertKeyLevelLevel2Ecdsa:
		keyId, err := strconv.ParseInt(info.CertKeyId, 10, 64)
		if err != nil {
			panic(err)
		}
		keyset, err := storage.LoadLevel2KeySetById(keyId)
		if err != nil {
			panic(err)
		}
		pri, _, err := keyset.Ecdsa()
		if err != nil {
			panic(err)
		}
		return pri
	case secretapi.CertKeyLevelLevel2Rsa:
		keyId, err := strconv.ParseInt(info.CertKeyId, 10, 64)
		if err != nil {
			panic(err)
		}
		keyset, err := storage.LoadLevel2KeySetById(keyId)
		if err != nil {
			panic(err)
		}
		pri, err := keyset.Rsa()
		if err != nil {
			panic(err)
		}
		return pri
	case secretapi.CertKeyLevelLevel2Custom:
		keyId, err := strconv.ParseInt(info.CertKeyId, 10, 64)
		if err != nil {
			panic(err)
		}
		keyType, key, err := storage.LoadL2DataKeyById(keyId)
		if err != nil {
			panic(err)
		}
		pri, err := secretapi.NewRawCipherTool().ConvertPKISupportedKey(keyType, key)
		if err != nil {
			panic(err)
		}
		return pri
	default:
		panic(errors.New("unknown certificate key level detected"))
	}
}

func initializeKeyStorage(pgConnStr string) SecretStorage {
	up, err := secret.NewLocalFileUnsealProvider(os.DirFS("."), map[int64]string{
		1: "bootstrap.key",
	})
	if err != nil {
		panic(err)
	}
	keyStorage, err := secretimpl.NewPostgresKeyStorage(pgConnStr)
	if err != nil {
		panic(err)
	}
	if err := keyStorage.SetupUnsealProviderAndWait(up); err != nil {
		panic(err)
	}
	log.Println("Secret unsealed success!")
	return keyStorage
}

func checkNonEmptyString(str string) {
	if len(str) <= 0 {
		panic(errors.New("empty string"))
	}
}

func getEnvString(c *configclient.EnvClient, key string) string {
	val, err := c.GetString(key)
	if err != nil {
		panic(err)
	}
	return val
}
