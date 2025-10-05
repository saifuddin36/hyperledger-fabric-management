
Blockchain Asset Management System (Hyperledger Fabric)

This project implements a secure, transparent, and immutable asset management system using Hyperledger Fabric. The solution includes a Go-based chaincode (smart contract) deployed on a two-organization test network and a containerized Node.js REST API that serves as the client application, interacting with the blockchain via the Fabric Gateway SDK.

Key Features
Asset Creation: Initialize new accounts (assets) on the ledger.
Asset Query: Read the current state of any asset from the World State.
Asset Update: Execute financial transactions (DEPOSIT/WITHDRAW) to update the asset balance.
Transaction History: Retrieve the full, immutable transaction history for any asset.
Project Structure
The repository is structured to separate the blockchain components (chaincode) from the client layer (REST API).

asset-management-system/
├── chaincode/
│ └── asset-manager/
│ ├── asset-chaincode.go # Go Smart Contract implementation
│ └── go.mod
└── rest-api/
├── app.js # Node.js Express API (Fabric Gateway SDK)
├── package.json # API dependencies
├── Dockerfile # Containerization for the API
└── crypto/ # Runtime directory for Fabric connection configs

Prerequisites
Before running the system, ensure you have the following installed:

Git: For cloning the repository.
Docker & Docker Compose: Essential for running Hyperledger Fabric and the REST API container.
Go (v1.20+): Required to build the chaincode.
Node.js (v18+): Required for the REST API.
Hyperledger Fabric Samples: The runbook assumes you are executing commands from within a cloned fabric-samples directory.
Runbook: Setup, Deployment, and Testing
This guide assumes you start from the root of the fabric-samples directory.

Step 1: Start the Hyperledger Fabric Network
Navigate to the test-network directory and bring up the network, creating the channel mychannel.

cd fabric-samples/test-network

# Start network and create channel 'mychannel'
./network.sh up createChannel -c mychannel

Step 2: Deploy the Chaincode
Package, install, approve, and commit the asset-manager chaincode to the channel.

# Ensure you are still in fabric-samples/test-network

# 1. Package the chaincode
peer lifecycle chaincode package asset_manager.tar.gz --path ../asset-management-system/chaincode/asset-manager/ --lang golang --label asset_manager_1

# 2. Install, Approve, and Commit (using the provided script for simplicity)
./network.sh deployCC -ccn asset_manager -ccp ../asset-management-system/chaincode/asset-manager -ccl go -ccv 1.0 -ccs 1 -c mychannel

# 3. Initialize the ledger (calls InitLedger)
./network.sh deployCC -ccn asset_manager -ccv 1.0 -ccs 1 -c mychannel -cci InitLedger -ccf '[]'

Step 3: Configure and Run the REST API
The Node.js application requires authentication materials to connect to the Fabric Gateway peer.

# Navigate to the REST API directory
cd ../../asset-management-system/rest-api

# 1. Copy necessary crypto materials (Admin identity and TLS certs for Org1)
# This creates the 'crypto' directory expected by the Docker container
mkdir -p crypto/peer/admin/msp/signcerts
mkdir -p crypto/peer/admin/msp/keystore
mkdir -p crypto/peer/tlscacerts

cp ../../test-network/organizations/peerOrganizations/[org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem](https://org1.example.com/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem) crypto/peer/admin/msp/signcerts/cert.pem
cp ../../test-network/organizations/peerOrganizations/[org1.example.com/users/Admin@org1.example.com/msp/keystore/*_sk](https://org1.example.com/users/Admin@org1.example.com/msp/keystore/\*\_sk) crypto/peer/admin/msp/keystore/key.pem
cp ../../test-network/organizations/peerOrganizations/[org1.example.com/peers/peer0.org1.example.com/tls/ca.crt](https://org1.example.com/peers/peer0.org1.example.com/tls/ca.crt) crypto/peer/tlscacerts/tlsca.pem

# 2. Build the API Docker image
docker build -t asset-management-api .

# 3. Run the container, connecting it to the test-network's Docker network
NETWORK_NAME=$(docker network ls -f name=net_test -q)
docker run -d \
--name asset-api \
--network $NETWORK_NAME \
-v $(pwd)/crypto:/etc/hyperledger/client \
-p 3000:3000 \
asset-management-api

API Usage and Testing
The API is now running on http://localhost:3000. Use curl or a tool like Postman to test the endpoints.

Chaincode Name: asset_manager
Channel Name: mychannel

1. Create Asset (Transaction)
Endpoint: POST /api/assets

curl -X POST http://localhost:3000/api/assets \
-H "Content-Type: application/json" \
-d '{
"DEALERID": "D003",
"MSISDN": "9955554444",
"MPIN": "9999",
"BALANCE": 100.00,
"STATUS": "Pending"
}'

2. Read Asset (Query World State)
Endpoint: GET /api/assets/:id

# Check the newly created asset
curl -X GET http://localhost:3000/api/assets/9955554444

3. Update Asset (Transaction)
Endpoint: PUT /api/assets/:id/transaction

This example performs a DEPOSIT of $500. TRANSTYPE also supports "WITHDRAW".

curl -X PUT http://localhost:3000/api/assets/9912345678/transaction \
-H "Content-Type: application/json" \
-d '{
"TRANSAMOUNT": 500.00,
"TRANSTYPE": "DEPOSIT",
"REMARKS": "Cash deposit via REST API"
}'

4. Retrieve Transaction History (Query)
Endpoint: GET /api/assets/:id/history

# Retrieve the full history, including the InitLedger entry and the recent deposit
curl -X GET http://localhost:3000/api/assets/9912345678/history

Cleanup
To stop and remove all components:

# Navigate back to the test-network directory
cd fabric-samples/test-network

# Stop and remove the API container
docker stop asset-api
docker rm asset-api

# Shut down the Fabric network
./network.sh down

About
No description, website, or topics provided.
Resources
 Readme
 Activity
Stars
 0 stars
Watchers
 0 watching
Forks
 0 forks
Releases
No releases published
Create a new release
Packages
No packages published
Publish your first package
Footer
© 2025 GitHub, Inc.
Footer navigation
Terms
Privacy
Security
