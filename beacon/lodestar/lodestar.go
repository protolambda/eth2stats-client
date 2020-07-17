package lodestar

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/alethio/eth2stats-client/beacon/polling"

	"github.com/dghubble/sling"
	"github.com/sirupsen/logrus"

	"github.com/alethio/eth2stats-client/beacon"
	"github.com/alethio/eth2stats-client/types"
)

var log = logrus.WithField("module", "lodestar")

type LodestarHTTPClient struct {
	api    *sling.Sling
	client *http.Client
}

func (s *LodestarHTTPClient) GetVersion() (string, error) {
	path := "v1/node/version"
	type lodestarVersion struct {
		Data struct {
			Version string `json:"version,omitempty"`
		} `json:"data"`
	}
	resp := new(lodestarVersion)
	_, err := s.api.New().Get(path).ReceiveSuccess(resp)
	if err != nil {
		return "", err
	}
	return resp.Data.Version, nil
}

func (s *LodestarHTTPClient) GetGenesisTime() (int64, error) {
	path := "lodestar/genesis_time"
	genesisTime := int64(0)
	_, err := s.api.New().Get(path).ReceiveSuccess(&genesisTime)
	if err != nil {
		return 0, err
	}
	return genesisTime, nil
}

func (s *LodestarHTTPClient) GetPeerCount() (int64, error) {
	path := "v1/node/peers"
	type lodestarPeers struct {
		Data []struct{
			// omit everything, we only care about the peer count.
		} `json:"data"`
	}
	resp := new(lodestarPeers)
	_, err := s.api.New().Get(path).ReceiveSuccess(resp)
	if err != nil {
		return 0, err
	}
	return int64(len(resp.Data)), nil
}

func (s *LodestarHTTPClient) GetAttestationsInPoolCount() (int64, error) {
	return 0, beacon.NotImplemented
}

func (s *LodestarHTTPClient) GetSyncStatus() (bool, error) {
	path := fmt.Sprintf("v1/node/syncing")
	type lodestarSyncing struct {
		Data struct {
			SyncDistance string `json:"sync_distance,omitempty"`
		} `json:"data"`
	}
	resp := new(lodestarSyncing)
	_, err := s.api.New().Get(path).ReceiveSuccess(resp)
	if err != nil {
		return false, err
	}
	distance, err := strconv.ParseUint(resp.Data.SyncDistance, 0, 64)
	if err != nil {
		fmt.Println(err)
	}
	return distance > 0, nil
}

func (s *LodestarHTTPClient) GetChainHead() (*types.ChainHead, error) {
	path := fmt.Sprintf("lodestar/head")
	type chainHead struct {
		HeadSlot           string `json:"head_slot"`
		HeadBlockRoot      string `json:"head_block_root"`
		FinalizedSlot      string `json:"finalized_slot"`
		FinalizedBlockRoot string `json:"finalized_block_root"`
		JustifiedSlot      string `json:"justified_slot"`
		JustifiedBlockRoot string `json:"justified_block_root"`
	}
	head := new(chainHead)
	_, err := s.api.New().Get(path).ReceiveSuccess(head)
	if err != nil {
		return nil, err
	}
	headSlot, err := strconv.ParseUint(head.HeadSlot, 0, 64)
	if err != nil {
		// pre genesis this is empty, return a default
		zeroChainHead := types.ChainHead{
			HeadSlot:           0,
			HeadBlockRoot:      "0x0",
			FinalizedSlot:      0,
			FinalizedBlockRoot: "0x0",
			JustifiedSlot:      0,
			JustifiedBlockRoot: "0x0",
		}
		return &zeroChainHead, nil
	}
	finalizedSlot, err := strconv.ParseUint(head.FinalizedSlot, 0, 64)
	if err != nil {
		return nil, err
	}
	justifiedSlot, err := strconv.ParseUint(head.JustifiedSlot, 0, 64)
	if err != nil {
		return nil, err
	}
	typesChainHead := types.ChainHead{
		HeadSlot:           headSlot,
		HeadBlockRoot:      head.HeadBlockRoot,
		FinalizedSlot:      finalizedSlot,
		FinalizedBlockRoot: head.FinalizedBlockRoot,
		JustifiedSlot:      justifiedSlot,
		JustifiedBlockRoot: head.JustifiedBlockRoot,
	}
	return &typesChainHead, nil
}

func (c *LodestarHTTPClient) SubscribeChainHeads() (beacon.ChainHeadSubscription, error) {
	sub := polling.NewChainHeadClientPoller(c)
	go sub.Start()

	return sub, nil
}

func New(httpClient *http.Client, baseURL string) *LodestarHTTPClient {
	return &LodestarHTTPClient{
		api:    sling.New().Client(httpClient).Base(baseURL),
		client: httpClient,
	}
}
