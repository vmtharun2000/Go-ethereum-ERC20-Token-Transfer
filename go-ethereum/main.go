package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"contractIntegration/contractAbi"
)

func main() {
	// Connect to Ethereum client
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	// Load ERC20 contract ABI
	contractAbi, err := loadTokenContractABI()
	if err != nil {
		log.Fatal(err)
	}

	// Sender's private key (replace with sender's private key)
	privateKey, err := crypto.HexToECDSA("replace with private key")
	if err != nil {
		log.Fatal(err)
	}

	fromAddress, _ := privateKeyToPublicKey(privateKey)
	toAddress := common.HexToAddress("0x9c0369dB74DD8864521dE61077aB7df913D95829")
	tokenAddress := common.HexToAddress("0x7BFDC4aDc2f24D54F0Bd71ce2d24c4a6470C2d26")

	// Construct the data payload for the transfer function
	data, err := contractAbi.Pack("transfer", toAddress, big.NewInt(1000000000000000000)) // 1 Token (assuming 18 decimals)
	if err != nil {
		log.Fatal(err)
	}

	// Construct the transaction
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Estimate gas limit for the transaction
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:     fromAddress,
		To:       &tokenAddress,
		GasPrice: gasPrice,
		Value:    nil,
		Data:     data,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Add some buffer to the estimated gas limit
	gasLimit = gasLimit + 10000 // Adjust this buffer value as needed

	// Construct the transaction
	tx := types.NewTransaction(nonce, tokenAddress, big.NewInt(0), gasLimit, gasPrice, data)
	// Sign the transaction
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// Send the transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Transfer sent: %s\n", signedTx.Hash().Hex())
}

func privateKeyToPublicKey(privateKey *ecdsa.PrivateKey) (common.Address, *ecdsa.PublicKey) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	return crypto.PubkeyToAddress(*publicKeyECDSA), publicKeyECDSA
}

func loadTokenContractABI() (abi.ABI, error) {
	return abi.JSON(strings.NewReader(contractAbi.Erc20ABI))
}
