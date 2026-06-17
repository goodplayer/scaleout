package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/meidoworks/nekoq-component/configure/secret/api"
	"github.com/meidoworks/nekoq-component/configure/secret/impl"
	"github.com/meidoworks/nekoq-component/configure/secret/tools"
	"github.com/meidoworks/nekoq-component/configure/secret/utils"
	"github.com/pelletier/go-toml/v2"

	"github.com/goodplayer/scaleout/consts"
)

type InitSrvConfig struct {
	PostgresConnectionString string           `toml:"postgres-connection-string"`
	RootCA                   CACertConfig     `toml:"root-ca"`
	IntermediateCA           CACertConfig     `toml:"intermediate-ca"`
	ClusterTLS               ClusterTLSConfig `toml:"cluster-tls"`
}

type CACertConfig struct {
	CommonName string `toml:"common-name"`
	Org        string `toml:"org"`
	Country    string `toml:"country"`
	Province   string `toml:"province"`
	Locality   string `toml:"locality"`
	Street     string `toml:"street"`
	Postal     string `toml:"postal"`
	Years      int    `toml:"years"`
}

type ClusterTLSConfig struct {
	Org      string `toml:"org"`
	Country  string `toml:"country"`
	Province string `toml:"province"`
	Locality string `toml:"locality"`
	Street   string `toml:"street"`
	Postal   string `toml:"postal"`
	Years    int    `toml:"years"`
	DNSNames string `toml:"dns-names"`
}

var (
	configFilePath string
)

func init() {
	flag.StringVar(&configFilePath, "config", "init_template.toml", "config file path")
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
	cfg := new(InitSrvConfig)
	err := toml.Unmarshal(fn(), cfg)
	if err != nil {
		panic(err)
	}

	pgConnStr := cfg.PostgresConnectionString
	checkNonEmptyString(pgConnStr)
	fmt.Println("debug pgConnStr:", pgConnStr)

	clusterDnsNames := strings.Split(cfg.ClusterTLS.DNSNames, ",")
	if len(clusterDnsNames) == 0 {
		fmt.Println("cluster dns names is empty")
		return
	}

	fmt.Println("Root CA Certificate Common Name:", cfg.RootCA.CommonName)
	fmt.Println("Root CA Organization:", cfg.RootCA.Org)
	fmt.Println("Root CA Country:", cfg.RootCA.Country)
	fmt.Println("Root CA Province:", cfg.RootCA.Province)
	fmt.Println("Root CA Locality:", cfg.RootCA.Locality)
	fmt.Println("Root CA Street:", cfg.RootCA.Street)
	fmt.Println("Root CA Postal Code:", cfg.RootCA.Postal)
	fmt.Println("Root CA Years:", cfg.RootCA.Years)
	fmt.Println("Intermediate CA Common Name:", cfg.IntermediateCA.CommonName)
	fmt.Println("Intermediate CA Organization:", cfg.IntermediateCA.Org)
	fmt.Println("Intermediate CA Country:", cfg.IntermediateCA.Country)
	fmt.Println("Intermediate CA Province:", cfg.IntermediateCA.Province)
	fmt.Println("Intermediate CA Locality:", cfg.IntermediateCA.Locality)
	fmt.Println("Intermediate CA Street:", cfg.IntermediateCA.Street)
	fmt.Println("Intermediate CA Postal Code:", cfg.IntermediateCA.Postal)
	fmt.Println("Intermediate CA Years:", cfg.IntermediateCA.Years)
	fmt.Println("Cluster TLS Certificate Common Name:", clusterDnsNames[0])
	fmt.Println("Cluster TLS Organization:", cfg.ClusterTLS.Org)
	fmt.Println("Cluster TLS Country:", cfg.ClusterTLS.Country)
	fmt.Println("Cluster TLS Province:", cfg.ClusterTLS.Province)
	fmt.Println("Cluster TLS Locality:", cfg.ClusterTLS.Locality)
	fmt.Println("Cluster TLS Street:", cfg.ClusterTLS.Street)
	fmt.Println("Cluster TLS Postal Code:", cfg.ClusterTLS.Postal)
	fmt.Println("Cluster TLS Years:", cfg.ClusterTLS.Years)
	fmt.Println("Cluster TLS DNS Names:", clusterDnsNames)
	fmt.Print("Confirm the information?(y/n)")
	var inputChar string
	if _, err := fmt.Scan(&inputChar); err != nil {
		panic(err)
	}
	if inputChar != "y" {
		fmt.Println("Do not accept the information.")
		fmt.Println("Exiting...")
		return
	}

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

	// create keys for certs
	l1KeySet, err := api.DefaultKeyGen.GenerateVitalKeySet()
	if err != nil {
		panic(err)
	}
	l1pri, err := new(utils.PemTool).ParseECDSAPrivateKey(l1KeySet.ECDSA_P521)
	if err != nil {
		panic(err)
	}
	l2KeySet, err := api.DefaultKeyGen.GenerateVitalKeySet()
	if err != nil {
		panic(err)
	}
	l2pri, err := new(utils.PemTool).ParseECDSAPrivateKey(l2KeySet.ECDSA_P521)
	if err != nil {
		panic(err)
	}
	if err := keyStorage.StoreLevel1KeySet(consts.RootLevel1Key, l1KeySet); err != nil {
		panic(err)
	}
	l1KeyId, _, err := keyStorage.LoadLevel1KeySet(consts.RootLevel1Key) // use internal methods to retrieve L1 Key id
	if err != nil {
		panic(err)
	}
	if err := keyStorage.StoreLevel2KeySet(consts.RootLevel1Key, consts.RootLevel2Key, l2KeySet); err != nil {
		panic(err)
	}
	l2KeyId, _, err := keyStorage.FetchLevel2KeySet(consts.RootLevel2Key)
	if err != nil {
		panic(err)
	}
	rootCACertSn, err := keyStorage.NextCertSerialNumber()
	if err != nil {
		panic(err)
	}
	rootCACertSnBig, err := rootCACertSn.ToBigInt()
	if err != nil {
		panic(err)
	}
	intermediateCACertSn, err := keyStorage.NextCertSerialNumber()
	if err != nil {
		panic(err)
	}
	intermediateCACertSnBig, err := intermediateCACertSn.ToBigInt()
	if err != nil {
		panic(err)
	}
	certTool := new(utils.CertTool)
	// root ca
	rootCACert, err := certTool.CreateRootCACertificate((&utils.CACertReq{
		SerialNumber:  rootCACertSnBig,
		CommonName:    cfg.RootCA.CommonName,
		Organization:  cfg.RootCA.Org,
		Country:       cfg.RootCA.Country,
		Province:      cfg.RootCA.Province,
		Locality:      cfg.RootCA.Locality,
		StreetAddress: cfg.RootCA.Street,
		PostalCode:    cfg.RootCA.Postal,
		StartTime:     time.Now(),
	}).Duration(time.Duration(cfg.RootCA.Years)*365*24*time.Hour), new(utils.CertKeyPair).FromPrivateKey(l1pri))
	if err != nil {
		panic(err)
	}
	newRootCACertSn, err := keyStorage.SaveRootCA(consts.RootCACert, rootCACert, api.CertKeyInfo{
		CertKeyLevel: api.CertKeyLevelLevel1Ecdsa,
		CertKeyId:    fmt.Sprint(l1KeyId),
	})
	if err != nil {
		panic(err)
	}
	rootCACert, _, _, err = keyStorage.LoadCertById(newRootCACertSn)
	if err != nil {
		panic(err)
	}
	// intermediate ca
	intermediateCACert, err := certTool.CreateIntermediateCACertificate((&utils.CACertReq{
		SerialNumber:  intermediateCACertSnBig,
		CommonName:    cfg.IntermediateCA.CommonName,
		Organization:  cfg.IntermediateCA.Org,
		Country:       cfg.IntermediateCA.Country,
		Province:      cfg.IntermediateCA.Province,
		Locality:      cfg.IntermediateCA.Locality,
		StreetAddress: cfg.IntermediateCA.Street,
		PostalCode:    cfg.IntermediateCA.Postal,
		StartTime:     time.Now(),
	}).Duration(time.Duration(cfg.IntermediateCA.Years)*365*24*time.Hour), rootCACert, new(utils.CertKeyPair).FromPrivateKey(l1pri), new(utils.CertKeyPair).FromPrivateKey(l2pri))
	if err != nil {
		panic(err)
	}
	newIntermediateCACertSn, err := keyStorage.SaveIntermediateCA(consts.RootIntermediateCACert, newRootCACertSn, intermediateCACert, api.CertKeyInfo{
		CertKeyLevel: api.CertKeyLevelLevel2Ecdsa,
		CertKeyId:    fmt.Sprint(l2KeyId),
	})
	if err != nil {
		panic(err)
	}
	intermediateCACert, _, _, err = keyStorage.LoadCertById(newIntermediateCACertSn)
	if err != nil {
		panic(err)
	}
	// cluster tls cert
	certKey, err := api.DefaultKeyGen.ECDSA(api.KeyECDSA384)
	if err != nil {
		panic(err)
	}
	certPriKey, err := new(utils.PemTool).ParseECDSAPrivateKey(certKey)
	if err != nil {
		panic(err)
	}
	if err := keyStorage.StoreL2DataKey(consts.RootLevel1Key, consts.RootClusterTLSKey, api.KeyECDSA384, certKey); err != nil {
		panic(err)
	}
	certKeyId, _, _, err := keyStorage.FetchL2DataKey(consts.RootClusterTLSKey)
	if err != nil {
		panic(err)
	}
	certSn, err := keyStorage.NextCertSerialNumber()
	if err != nil {
		panic(err)
	}
	certSnBig, err := certSn.ToBigInt()
	if err != nil {
		panic(err)
	}
	certReq, err := certTool.CreateCertificateRequest(&utils.CertReq{
		CommonName:    clusterDnsNames[0],
		Organization:  cfg.ClusterTLS.Org,
		Country:       cfg.ClusterTLS.Country,
		Province:      cfg.ClusterTLS.Province,
		Locality:      cfg.ClusterTLS.Locality,
		StreetAddress: cfg.ClusterTLS.Street,
		PostalCode:    cfg.ClusterTLS.Postal,
		DNSNames:      clusterDnsNames,
	}, &utils.CertKeyPair{
		PrivateKey: certPriKey,
		PublicKey:  certPriKey.Public(),
	})
	if err != nil {
		panic(err)
	}
	clusterTLSCert, err := certTool.CreateCertificate(certReq, (&utils.CertMeta{
		SerialNumber: certSnBig,
		StartTime:    time.Now(),
		SignerCert:   intermediateCACert,
		Signer: &utils.CertKeyPair{
			PrivateKey: l2pri,
			PublicKey:  l2pri.Public(),
		},
	}).Duration(time.Duration(cfg.ClusterTLS.Years)*365*24*time.Hour))
	if err != nil {
		panic(err)
	}
	newCertSn, err := keyStorage.SaveCert(consts.RootClusterTLSCert, intermediateCACertSn, clusterTLSCert, api.CertKeyInfo{
		CertKeyLevel: api.CertKeyLevelLevel2Custom,
		CertKeyId:    fmt.Sprint(certKeyId),
	})
	if err != nil {
		panic(err)
	}

	// init jwt token key
	tool := tools.NewLevel2CipherTool(keyStorage, api.DefaultKeyGen, consts.RootLevel1Key)
	if err := tool.NewGeneral128BKey(consts.RootJwtTokenKey); err != nil {
		panic(err)
	}

	var certs = struct {
		RootCASn         api.CertSerialNumber
		IntermediateCASn api.CertSerialNumber
		CertSn           api.CertSerialNumber
		CertKeyId        int64
	}{RootCASn: newRootCACertSn, IntermediateCASn: newIntermediateCACertSn, CertSn: newCertSn, CertKeyId: certKeyId}
	fmt.Println("Created certificates:")
	fmt.Println("Root CA Cert Serial Number:", certs.RootCASn)
	fmt.Println("Intermediate CA Cert Serial Number:", certs.IntermediateCASn)
	fmt.Println("Cluster TLS Cert Serial Number:", certs.CertSn)
	fmt.Println("Cluster TLS Cert Signing Key Id:", certs.CertKeyId)

	// write certs to files
	if err := writeCertFileBySn("RootCa.crt", certs.RootCASn, keyStorage); err != nil {
		panic(err)
	}
	if err := writeCertFileBySn("IntermediateCa.crt", certs.IntermediateCASn, keyStorage); err != nil {
		panic(err)
	}
}

func writeCertFileBySn(name string, sn api.CertSerialNumber, keyStorage *impl.PostgresKeyStorage) error {
	cert, _, _, err := keyStorage.LoadCertById(sn)
	if err != nil {
		return err
	}
	data, err := new(utils.PemTool).EncodeCertificate(cert)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func checkNonEmptyString(str string) {
	if len(str) <= 0 {
		panic(errors.New("empty string"))
	}
}
