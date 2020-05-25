package main

import (
	"fmt"
	"github.com/go-acme/lego/challenge/dns01"
	"log"
	"net/http"
	"time"

	"github.com/foomo/simplecert"
	"github.com/foomo/tlsconfig"
)

type Handler struct{}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello from simplecert"))
}

func main() {

	myDNS, err := NewDNSProviderMyDNS()

	// do the cert magic
	cfg := simplecert.Default
	cfg.Domains = []string{"example.net", "*.example.net"}
	cfg.CacheDir = "letsencrypt"
	cfg.SSLEmail = "my@example.com"
	cfg.Local = false
	cfg.UpdateHosts = false
	//cfg.HTTPAddress = ":6666"
	//cfg.TLSAddress = ":6443"
	cfg.HTTPAddress = ""
	cfg.TLSAddress = ""
	//cfg.DNSProvider = "alidns"
	cfg.CustomProvider = myDNS

	certReloader, err := simplecert.Init(cfg, func() {
		// this function will be called upon receiving the syscall.SIGINT or syscall.SIGABRT signal
		// and can be used to stop your backend gracefully
		fmt.Println("cleaning up...")
	})
	if err != nil {
		log.Fatal("simplecert init failed: ", err)
	}

	// redirect HTTP to HTTPS
	log.Println("starting HTTP Listener on Port 80")
	go http.ListenAndServe(":80", http.HandlerFunc(simplecert.Redirect))

	// init strict tlsConfig with certReloader
	tlsconf := tlsconfig.NewServerTLSConfig(tlsconfig.TLSModeServerStrict)

	// now set GetCertificate to the reloaders GetCertificateFunc to enable hot reload
	tlsconf.GetCertificate = certReloader.GetCertificateFunc()

	// init server
	s := &http.Server{
		Addr:      ":443",
		TLSConfig: tlsconf,
		Handler:   Handler{},
	}

	log.Println("now visit: https://" + cfg.Domains[0])

	// lets go
	log.Fatal(s.ListenAndServeTLS("", ""))
}

type DNSProviderMyDNS struct {
	apiAuthToken string
}

func NewDNSProviderMyDNS() (*DNSProviderMyDNS, error) {
	return &DNSProviderMyDNS{}, nil
}

func (d *DNSProviderMyDNS) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	log.Println(fqdn, value)
	time.Sleep(30 * time.Second)
	// make API request to set a TXT record on fqdn with value and ttl
	return nil
}

func (d *DNSProviderMyDNS) CleanUp(domain, token, keyAuth string) error {
	// clean up any state you created in Present, like removing the TXT record
	log.Println("clean")
	return nil
}
