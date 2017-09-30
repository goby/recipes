package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/http2"
)

func init() {
	log.SetOutput(GlogWriter{})
}

// GlogWriter serves as a bridge between the standard log package and the glog package.
type GlogWriter struct{}

// Write implements the io.Writer interface.
func (writer GlogWriter) Write(data []byte) (n int, err error) {
	glog.Info(string(data))
	return len(data), nil
}

func main() {
	defer glog.Flush()

	var cafile, certfile, keyfile string
	flag.StringVar(&cafile, "ca", "", "ca file path")
	flag.StringVar(&certfile, "cert", "", "cert file path")
	flag.StringVar(&keyfile, "key", "", "key file path")
	flag.Parse()

	rootCA, err := ioutil.ReadFile(cafile)
	if err != nil {
		glog.Fatalf("Read cafile faild: %v", err)
	}

	certData, err := ioutil.ReadFile(certfile)
	if err != nil {
		glog.Fatalf("Read cert faild: %v", err)
	}

	keyData, err := ioutil.ReadFile(keyfile)
	if err != nil {
		glog.Fatalf("Read key faild: %v", err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(rootCA)
	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		glog.Fatalf("Create key pair failed: %v", err)
	}

	conf := &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{cert},
	}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		TLSClientConfig:    conf,
	}

	if err = http2.ConfigureTransport(tr); err != nil {
		glog.Fatalf("Config transport failed: %v", err)
	}

	client := &http.Client{Transport: tr, Timeout: 100 * time.Millisecond}
	_, _ = client.Get("https://10.166.224.124:6443")

	glog.V(1).Infof("===============START================")
	wg := sync.WaitGroup{}
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			client := &http.Client{Transport: tr, Timeout: 100 * time.Millisecond}
			_, err = client.Get("https://10.166.224.124:6443/api/v1/watch/namespaces/default2/services?fuck=true")

			if err != nil {
				glog.Errorf("==== Get response failed: %v", err)
			}

			time.Sleep(1 * time.Second)

			wg.Done()
		}()
	}

	wg.Wait()
}
