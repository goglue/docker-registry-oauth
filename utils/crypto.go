package utils

import (
	"crypto/rsa"
	"os"
	"io/ioutil"
	"encoding/pem"
	"crypto/x509"
)

func PrivateKey(file string) *rsa.PrivateKey {
	kf, err := os.Open(file)
	if nil != err {
		panic(err)
	}

	cfb, err := ioutil.ReadAll(kf)
	if nil != err {
		panic(err)
	}

	block, _ := pem.Decode(cfb)
	parseResult, _ := x509.ParsePKCS8PrivateKey(block.Bytes)
	key := parseResult.(*rsa.PrivateKey)

	return key
}

func Certificate(file string) *x509.Certificate {
	cf, err := os.Open(file)
	if nil != err {
		panic(err)
	}

	cfb, err := ioutil.ReadAll(cf)
	if nil != err {
		panic(err)
	}

	cpb, _ := pem.Decode(cfb)
	crt, err := x509.ParseCertificate(cpb.Bytes)
	if nil != err {
		panic(err)
	}

	return crt
}
