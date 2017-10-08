package insecure

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/philips/grpc-gateway-example/insecure"
)

const (
	port = 10000
)

var (
	demoKeyPair  *tls.Certificate
	demoCertPool *x509.CertPool
	demoAddr     string
)

// GetCert returns a certicicate pair and pool
func GetCert() (*tls.Certificate, *x509.CertPool) {
	var err error
	pair, err := tls.X509KeyPair([]byte(insecure.Cert), []byte(insecure.Key))
	if err != nil {
		panic(err)
	}
	demoKeyPair = &pair
	demoCertPool = x509.NewCertPool()
	ok := demoCertPool.AppendCertsFromPEM([]byte(insecure.Cert))
	if !ok {
		panic("bad certs")
	}
	demoAddr = fmt.Sprintf("localhost:%d", port)

	return demoKeyPair, demoCertPool
}
