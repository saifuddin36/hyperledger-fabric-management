package main

import (
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID = "Org1MSP"
	// This path points to the organizations folder we copied into our project
	cryptoPath  = "organizations/peerOrganizations/org1.example.com"
	certPath    = cryptoPath + "/users/User1@org1.example.com/msp/signcerts/cert.pem"
	keyPath     = cryptoPath + "/users/User1@org1.example.com/msp/keystore/"
	tlsCertPath = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	// IMPORTANT: This now uses the Docker container name instead of "localhost"
	peerEndpoint = "peer0.org1.example.com:7051"
	gatewayPeer  = "peer0.org1.example.com"
)

var contract *client.Contract

func main() {
	log.Println("============ Application starts ============")

	clientConnection, err := newGrpcConnection()
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}
	defer clientConnection.Close()

	id := newIdentity()
	sign := newSign()

	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(fmt.Errorf("failed to connect to gateway: %w", err))
	}
	defer gw.Close()

	network := gw.GetNetwork("mychannel")
	contract = network.GetContract("account")

	r := mux.NewRouter()
	r.HandleFunc("/accounts/{msisdn}", queryAccountHandler).Methods("GET")

	log.Println("Starting server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func queryAccountHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	msisdn := vars["msisdn"]

	log.Printf("--> Evaluate Transaction: ReadAccount, function returns account for %s\n", msisdn)

	evaluateResult, err := contract.EvaluateTransaction("ReadAccount", msisdn)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to evaluate transaction: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(evaluateResult)
}

func newGrpcConnection() (*grpc.ClientConn, error) {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	return grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
}

func newIdentity() *identity.X509Identity {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(path string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

func newSign() identity.Sign {
	files, err := os.ReadDir(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key directory: %w", err))
	}
	privateKeyPEM, err := os.ReadFile(path.Join(keyPath, files[0].Name()))
	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}