package main

import (
	"errors"
	"fmt"
	"sync"

	"github.com/taurusgroup/multi-party-sig/pkg/ecdsa"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/pool"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
	"github.com/taurusgroup/multi-party-sig/protocols/example"
)

func XOR(id party.ID, ids party.IDSlice, n *Network) error {
	h, err := protocol.NewMultiHandler(example.StartXOR(id, ids), nil)
	if err != nil {
		return err
	}
	HandlerLoop(id, h, n)
	_, err = h.Result()
	if err != nil {
		return err
	}
	return nil
}

func CMPKeygen(id party.ID, ids party.IDSlice, threshold int, n *Network, pl *pool.Pool) (*cmp.Config, error) {
	h, err := protocol.NewMultiHandler(cmp.Keygen(curve.Secp256k1{}, id, ids, threshold, pl), nil)
	if err != nil {
		return nil, err
	}
	HandlerLoop(id, h, n)
	r, err := h.Result()
	if err != nil {
		return nil, err
	}

	return r.(*cmp.Config), nil
}

func CMPRefresh(c *cmp.Config, n *Network, pl *pool.Pool) (*cmp.Config, error) {
	hRefresh, err := protocol.NewMultiHandler(cmp.Refresh(c, pl), nil)
	if err != nil {
		return nil, err
	}
	HandlerLoop(c.ID, hRefresh, n)

	r, err := hRefresh.Result()
	if err != nil {
		return nil, err
	}

	return r.(*cmp.Config), nil
}

func CMPSign(c *cmp.Config, m []byte, signers party.IDSlice, n *Network, pl *pool.Pool) error {
	h, err := protocol.NewMultiHandler(cmp.Sign(c, signers, m, pl), nil)
	if err != nil {
		return err
	}
	HandlerLoop(c.ID, h, n)

	signResult, err := h.Result()
	if err != nil {
		return err
	}
	signature := signResult.(*ecdsa.Signature)
	if !signature.Verify(c.PublicPoint(), m) {
		return errors.New("failed to verify cmp signature")
	}
	return nil
}

func CMPPreSign(c *cmp.Config, signers party.IDSlice, n *Network, pl *pool.Pool) (*ecdsa.PreSignature, error) {
	h, err := protocol.NewMultiHandler(cmp.Presign(c, signers, pl), nil)
	if err != nil {
		return nil, err
	}

	HandlerLoop(c.ID, h, n)

	signResult, err := h.Result()
	if err != nil {
		return nil, err
	}

	preSignature := signResult.(*ecdsa.PreSignature)
	if err = preSignature.Validate(); err != nil {
		return nil, errors.New("failed to verify cmp presignature")
	}
	return preSignature, nil
}

func CMPPreSignOnline(c *cmp.Config, preSignature *ecdsa.PreSignature, m []byte, n *Network, pl *pool.Pool) error {
	h, err := protocol.NewMultiHandler(cmp.PresignOnline(c, preSignature, m, pl), nil)
	if err != nil {
		return err
	}
	HandlerLoop(c.ID, h, n)

	signResult, err := h.Result()
	if err != nil {
		return err
	}
	signature := signResult.(*ecdsa.Signature)
	fmt.Println(signature)
	fmt.Println(c.PublicPoint())
	if !signature.Verify(c.PublicPoint(), m) {
		return errors.New("failed to verify cmp signature")
	}
	return nil
}

func All(id party.ID, ids party.IDSlice, threshold int, message []byte, n *Network, wg *sync.WaitGroup, pl *pool.Pool) error {
	defer wg.Done()

	// XOR
	err := XOR(id, ids, n)
	if err != nil {
		return err
	}

	// CMP KEYGEN
	keygenConfig, err := CMPKeygen(id, ids, threshold, n, pl)
	if err != nil {
		return err
	}

	// CMP REFRESH
	refreshConfig, err := CMPRefresh(keygenConfig, n, pl)
	if err != nil {
		return err
	}

	signers := ids[:threshold+1]
	if !signers.Contains(id) {
		n.Quit(id)
		return nil
	}

	// CMP SIGN
	err = CMPSign(refreshConfig, message, signers, n, pl)
	if err != nil {
		return err
	}

	// CMP PRESIGN
	preSignature, err := CMPPreSign(refreshConfig, signers, n, pl)
	if err != nil {
		return err
	}

	// CMP PRESIGN ONLINE
	err = CMPPreSignOnline(refreshConfig, preSignature, message, n, pl)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	ids := party.IDSlice{"a", "b", "c", "d", "e", "f"}
	threshold := 4
	messageToSign := []byte("hello")

	net := NewNetwork(ids)

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id party.ID) {
			pl := pool.NewPool(0)
			defer pl.TearDown()
			if err := All(id, ids, threshold, messageToSign, net, &wg, pl); err != nil {
				fmt.Println(err)
			}
		}(id)
	}
	wg.Wait()
}
