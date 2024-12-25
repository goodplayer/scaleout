package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/meidoworks/nekoq-component/configure/configclient"
	"github.com/meidoworks/nekoq-component/configure/secret"
	"github.com/meidoworks/nekoq-component/configure/secretapi"
	"github.com/meidoworks/nekoq-component/configure/secretimpl"

	"github.com/goodplayer/scaleout/consts"
)

var (
	rootCaCommonName string
	rootCaOrg        string
	rootCaCountry    string
	rootCaProvince   string
	rootCaLocality   string
	rootCaStreet     string
	rootCaPostal     string
	rootCaYears      int
)

var (
	intermediateCaCommonName string
	intermediateCaOrg        string
	intermediateCaCountry    string
	intermediateCaProvince   string
	intermediateCaLocality   string
	intermediateCaStreet     string
	intermediateCaPostal     string
	intermediateCaYears      int
)

var (
	clusterTlsCertOrg      string
	clusterTlsCertCountry  string
	clusterTlsCertProvince string
	clusterTlsCertLocality string
	clusterTlsCertStreet   string
	clusterTlsCertPostal   string
	clusterTlsCertYears    int
	clusterTlsCertDnsNames string
)

func init() {
	flag.StringVar(&rootCaCommonName, "root-ca-common-name", "Test Root CA", "root CA common name")
	flag.StringVar(&rootCaOrg, "root-ca-org", "Test Organization", "root CA organization")
	flag.StringVar(&rootCaCountry, "root-ca-country", "CN", "root CA country")
	flag.StringVar(&rootCaProvince, "root-ca-province", "Test Province", "root CA province")
	flag.StringVar(&rootCaLocality, "root-ca-locality", "Test Locality", "root CA locality")
	flag.StringVar(&rootCaStreet, "root-ca-street", "Test Street", "root CA street")
	flag.StringVar(&rootCaPostal, "root-ca-postal", "Test Postal Code", "root CA postal")
	flag.IntVar(&rootCaYears, "root-ca-years", 30, "root CA years")

	flag.StringVar(&intermediateCaCommonName, "intermediate-ca-common-name", "Test Intermediate CA", "intermediate CA common name")
	flag.StringVar(&intermediateCaOrg, "intermediate-ca-org", "Test Organization", "intermediate CA organization")
	flag.StringVar(&intermediateCaCountry, "intermediate-ca-country", "CN", "intermediate CA country")
	flag.StringVar(&intermediateCaProvince, "intermediate-ca-province", "Test Province", "intermediate CA province")
	flag.StringVar(&intermediateCaLocality, "intermediate-ca-locality", "Test Locality", "intermediate CA locality")
	flag.StringVar(&intermediateCaStreet, "intermediate-ca-street", "Test Street", "intermediate CA street")
	flag.StringVar(&intermediateCaPostal, "intermediate-ca-postal", "Test Postal Code", "intermediate CA postal")
	flag.IntVar(&intermediateCaYears, "intermediate-ca-years", 15, "intermediate CA years")

	flag.StringVar(&clusterTlsCertOrg, "cluster-tls-org", "Test Organization", "cluster TLS organization")
	flag.StringVar(&clusterTlsCertCountry, "cluster-tls-country", "CN", "cluster TLS country")
	flag.StringVar(&clusterTlsCertProvince, "cluster-tls-province", "Test Province", "cluster TLS province")
	flag.StringVar(&clusterTlsCertLocality, "cluster-tls-locality", "Test Locality", "cluster TLS locality")
	flag.StringVar(&clusterTlsCertStreet, "cluster-tls-street", "Test Street", "cluster TLS street")
	flag.StringVar(&clusterTlsCertPostal, "cluster-tls-postal", "Test Postal Code", "cluster TLS postal")
	flag.IntVar(&clusterTlsCertYears, "cluster-tls-years", 10, "cluster TLS years")
	flag.StringVar(&clusterTlsCertDnsNames, "cluster-tls-dns-names", "localhost,127.0.0.1", "cluster TLS dns names")

	flag.Parse()
}

func main() {
	cfgclient := configclient.NewEnvClient()
	pgConnStr := getEnvString(cfgclient, "POSTGRES_CONNECTION_STRING")
	checkNonEmptyString(pgConnStr)
	fmt.Println("debug pgConnStr:", pgConnStr)

	clusterDnsNames := strings.Split(clusterTlsCertDnsNames, ",")
	if len(clusterDnsNames) == 0 {
		fmt.Println("cluster dns names is empty")
		return
	}

	fmt.Println("Root CA Certificate Common Name:", rootCaCommonName)
	fmt.Println("Root CA Organization:", rootCaOrg)
	fmt.Println("Root CA Country:", rootCaCountry)
	fmt.Println("Root CA Province:", rootCaProvince)
	fmt.Println("Root CA Locality:", rootCaLocality)
	fmt.Println("Root CA Street:", rootCaStreet)
	fmt.Println("Root CA Postal Code:", rootCaPostal)
	fmt.Println("Root CA Years:", rootCaYears)
	fmt.Println("Intermediate CA Common Name:", intermediateCaCommonName)
	fmt.Println("Intermediate CA Organization:", intermediateCaOrg)
	fmt.Println("Intermediate CA Country:", intermediateCaCountry)
	fmt.Println("Intermediate CA Province:", intermediateCaProvince)
	fmt.Println("Intermediate CA Locality:", intermediateCaLocality)
	fmt.Println("Intermediate CA Street:", intermediateCaStreet)
	fmt.Println("Intermediate CA Postal Code:", intermediateCaPostal)
	fmt.Println("Intermediate CA Years:", intermediateCaYears)
	fmt.Println("Cluster TLS Certificate Common Name:", clusterDnsNames[0])
	fmt.Println("Cluster TLS Organization:", clusterTlsCertOrg)
	fmt.Println("Cluster TLS Country:", clusterTlsCertCountry)
	fmt.Println("Cluster TLS Province:", clusterTlsCertProvince)
	fmt.Println("Cluster TLS Locality:", clusterTlsCertLocality)
	fmt.Println("Cluster TLS Street:", clusterTlsCertStreet)
	fmt.Println("Cluster TLS Postal Code:", clusterTlsCertPostal)
	fmt.Println("Cluster TLS Years:", clusterTlsCertYears)
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
	if err := keyStorage.Startup(); err != nil {
		panic(err)
	}
	if err := keyStorage.SetupUnsealProviderAndWait(up); err != nil {
		panic(err)
	}
	fmt.Println("unseal success!")

	// create keys for certs
	l1KeySet, err := secretapi.DefaultKeyGen.GenerateVitalKeySet()
	if err != nil {
		panic(err)
	}
	l1pri, err := new(secretapi.PemTool).ParseECDSAPrivateKey(l1KeySet.ECDSA_P521)
	if err != nil {
		panic(err)
	}
	l2KeySet, err := secretapi.DefaultKeyGen.GenerateVitalKeySet()
	if err != nil {
		panic(err)
	}
	l2pri, err := new(secretapi.PemTool).ParseECDSAPrivateKey(l2KeySet.ECDSA_P521)
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
	certTool := new(secretapi.CertTool)
	// root ca
	rootCACert, err := certTool.CreateRootCACertificate((&secretapi.CACertReq{
		SerialNumber:  rootCACertSnBig,
		CommonName:    rootCaCommonName,
		Organization:  rootCaOrg,
		Country:       rootCaCountry,
		Province:      rootCaProvince,
		Locality:      rootCaLocality,
		StreetAddress: rootCaStreet,
		PostalCode:    rootCaPostal,
		StartTime:     time.Now(),
	}).Duration(time.Duration(rootCaYears)*365*24*time.Hour), new(secretapi.CertKeyPair).FromPrivateKey(l1pri))
	if err != nil {
		panic(err)
	}
	newRootCACertSn, err := keyStorage.SaveRootCA(consts.RootCACert, rootCACert, secretapi.CertKeyInfo{
		CertKeyLevel: secretapi.CertKeyLevelLevel1Ecdsa,
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
	intermediateCACert, err := certTool.CreateIntermediateCACertificate((&secretapi.CACertReq{
		SerialNumber:  intermediateCACertSnBig,
		CommonName:    intermediateCaCommonName,
		Organization:  intermediateCaOrg,
		Country:       intermediateCaCountry,
		Province:      intermediateCaProvince,
		Locality:      intermediateCaLocality,
		StreetAddress: intermediateCaStreet,
		PostalCode:    intermediateCaPostal,
		StartTime:     time.Now(),
	}).Duration(time.Duration(intermediateCaYears)*365*24*time.Hour), rootCACert, new(secretapi.CertKeyPair).FromPrivateKey(l1pri), new(secretapi.CertKeyPair).FromPrivateKey(l2pri))
	if err != nil {
		panic(err)
	}
	newIntermediateCACertSn, err := keyStorage.SaveIntermediateCA(consts.RootIntermediateCACert, newRootCACertSn, intermediateCACert, secretapi.CertKeyInfo{
		CertKeyLevel: secretapi.CertKeyLevelLevel2Ecdsa,
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
	certKey, err := secretapi.DefaultKeyGen.ECDSA(secretapi.KeyECDSA384)
	if err != nil {
		panic(err)
	}
	certPriKey, err := new(secretapi.PemTool).ParseECDSAPrivateKey(certKey)
	if err != nil {
		panic(err)
	}
	if err := keyStorage.StoreL2DataKey(consts.RootLevel1Key, consts.RootClusterTLSKey, secretapi.KeyECDSA384, certKey); err != nil {
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
	certReq, err := certTool.CreateCertificateRequest(&secretapi.CertReq{
		CommonName:    clusterDnsNames[0],
		Organization:  clusterTlsCertOrg,
		Country:       clusterTlsCertCountry,
		Province:      clusterTlsCertProvince,
		Locality:      clusterTlsCertLocality,
		StreetAddress: clusterTlsCertStreet,
		PostalCode:    clusterTlsCertPostal,
		DNSNames:      clusterDnsNames,
	}, &secretapi.CertKeyPair{
		PrivateKey: certPriKey,
		PublicKey:  certPriKey.Public(),
	})
	if err != nil {
		panic(err)
	}
	clusterTLSCert, err := certTool.CreateCertificate(certReq, (&secretapi.CertMeta{
		SerialNumber: certSnBig,
		StartTime:    time.Now(),
		SignerCert:   intermediateCACert,
		Signer: &secretapi.CertKeyPair{
			PrivateKey: l2pri,
			PublicKey:  l2pri.Public(),
		},
	}).Duration(time.Duration(clusterTlsCertYears)*365*24*time.Hour))
	if err != nil {
		panic(err)
	}
	newCertSn, err := keyStorage.SaveCert(consts.RootClusterTLSCert, intermediateCACertSn, clusterTLSCert, secretapi.CertKeyInfo{
		CertKeyLevel: secretapi.CertKeyLevelLevel2Custom,
		CertKeyId:    fmt.Sprint(certKeyId),
	})
	if err != nil {
		panic(err)
	}

	// init jwt token key
	tool := secretapi.NewLevel2CipherTool(keyStorage, secretapi.DefaultKeyGen, consts.RootLevel1Key)
	if err := tool.NewGeneral128BKey(consts.RootJwtTokenKey); err != nil {
		panic(err)
	}

	var certs = struct {
		RootCASn         secretapi.CertSerialNumber
		IntermediateCASn secretapi.CertSerialNumber
		CertSn           secretapi.CertSerialNumber
		CertKeyId        int64
	}{RootCASn: newRootCACertSn, IntermediateCASn: newIntermediateCACertSn, CertSn: newCertSn, CertKeyId: certKeyId}
	fmt.Println("Created certificates:")
	fmt.Println("Root CA Cert Serial Number:", certs.RootCASn)
	fmt.Println("Intermediate CA Cert Serial Number:", certs.IntermediateCASn)
	fmt.Println("Cluster TLS Cert Serial Number:", certs.CertSn)
	fmt.Println("Cluster TLS Cert Signing Key Id:", certs.CertKeyId)

	// write certs to files
	if err := writeCertFileBySn("rootCa.crt", certs.RootCASn, keyStorage); err != nil {
		panic(err)
	}
	if err := writeCertFileBySn("intermediateCa.crt", certs.IntermediateCASn, keyStorage); err != nil {
		panic(err)
	}
}

func writeCertFileBySn(name string, sn secretapi.CertSerialNumber, keyStorage *secretimpl.PostgresKeyStorage) error {
	cert, _, _, err := keyStorage.LoadCertById(sn)
	if err != nil {
		return err
	}
	data, err := new(secretapi.PemTool).EncodeCertificate(cert)
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
