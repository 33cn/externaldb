package main

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/33cn/externaldb/db/multisig"
	"github.com/33cn/externaldb/escli/querypara"
)

// MultiSig MultiSig
type MultiSig struct {
	*DBRead
}

func decodeAllMS(x *json.RawMessage) (interface{}, error) {
	/* x可能是
	 1. multisig.MS{}
	 2. multisig.MSLimit{}
	 3. multisig.MSOwner{}
	把其中某一类型decode成另一个类型json.Unmarshal不会报错,只会对共同的字段进行赋值
	*/
	r, err := decodeMS(x)
	if err != nil {
		return nil, err
	}
	ms := r.(*multisig.MS)
	switch ms.Type {
	case multisig.MSTypeAccountX:
		return r, nil
	case multisig.MSTypeLimitX:
		return decodeMSLimit(x)
	case multisig.MSTypeOwnerX:
		return decodeMSOwner(x)
	}

	return nil, errors.New("error multiSig type")

}

func decodeMS(x *json.RawMessage) (interface{}, error) {
	r := multisig.MS{}
	err := json.Unmarshal([]byte(*x), &r)
	return &r, err
}

func decodeMSLimit(x *json.RawMessage) (interface{}, error) {
	r := multisig.MSLimit{}
	err := json.Unmarshal([]byte(*x), &r)
	return &r, err
}

func decodeMSOwner(x *json.RawMessage) (interface{}, error) {
	r := multisig.MSOwner{}
	err := json.Unmarshal([]byte(*x), &r)
	return &r, err
}

// OutMS OutMS
type OutMS struct {
	*multisig.MS
	Limits []*multisig.MSLimit `json:"limits"`
	Owners []*multisig.MSOwner `json:"owners"`
}

func mapAddress(ms []interface{}) (map[string]*OutMS, error) {
	addresses := make(map[string]*OutMS)
	for _, m := range ms {
		x, ok := m.(*multisig.MS)
		if !ok {
			return nil, errors.New(ErrTypeAsset)
		}
		addresses[x.MultiSigAddr] = &OutMS{MS: x, Limits: []*multisig.MSLimit{}, Owners: []*multisig.MSOwner{}}
	}
	return addresses, nil
}

func getAddress(ms []interface{}) ([]string, error) {
	addresses := make(map[string]bool)

	for _, m := range ms {
		x, ok := m.(*multisig.MS)
		if ok {
			addresses[x.MultiSigAddr] = true
			continue
		}
		l, ok := m.(*multisig.MSLimit)
		if ok {
			addresses[l.MultiSigAddr] = true
			continue
		}
		o, ok := m.(*multisig.MSOwner)
		if ok {
			addresses[o.MultiSigAddr] = true
			continue
		}
		return nil, errors.New(ErrTypeAsset)
	}

	out := make([]string, 0)
	for k := range addresses {
		out = append(out, k)
	}
	return out, nil
}

func getAddresses(resp [][]interface{}) ([][]string, error) {
	addresses := make([][]string, 0)
	for _, ms := range resp {
		addr, err := getAddress(ms)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, addr)
	}

	return addresses, nil
}

// Gets gets
func (b *MultiSig) Gets(req *Addresses, out *interface{}) error {
	if req == nil || len(req.Address) == 0 {
		return nil
	}
	ids := make([]string, 0)
	for _, addr := range req.Address {
		s := multisig.MS{MultiSigAddr: addr}
		ids = append(ids, s.ID())
	}

	resp, err := b.DBRead.gets(multisig.MSDBX, multisig.MSDBX, ids, decodeMS)
	if err != nil {
		return err
	}

	outMS, err := mapAddress(resp)
	if err != nil {
		return err
	}

	err = b.loadLimits(outMS)
	if err != nil {
		return err
	}

	err = b.loadOwners(outMS)
	if err != nil {
		return err
	}
	ms := make([]*OutMS, 0)
	for _, v := range outMS {
		ms = append(ms, v)
	}

	*out = ms
	return err
}

func (b *MultiSig) loadLimits(outMS map[string]*OutMS) error {
	qs := make([]*querypara.Query, 0)
	for k := range outMS {
		q := &querypara.Query{
			Page: &querypara.QPage{
				Size:   1000,
				Number: 1,
			},
			Match: []*querypara.QMatch{
				{Key: "multi_signature_address", Value: k},
				{Key: "type", Value: multisig.MSTypeLimitX},
			},
		}

		qs = append(qs, q)
	}
	resp, err := b.DBRead.searches(multisig.MSDBX, multisig.MSDBX, qs, decodeMSLimit)
	if err != nil {
		return err
	}
	for _, rs := range resp {
		for _, r := range rs {
			l, ok := r.(*multisig.MSLimit)
			if !ok {
				return errors.New(ErrTypeAsset)
			}
			outMS[l.MultiSigAddr].Limits = append(outMS[l.MultiSigAddr].Limits, l)
		}
	}
	return nil
}

func (b *MultiSig) loadOwners(outMS map[string]*OutMS) error {
	qs := make([]*querypara.Query, 0)
	for k := range outMS {
		q := &querypara.Query{
			Page: &querypara.QPage{
				Size:   1000,
				Number: 1,
			},
			Match: []*querypara.QMatch{
				{Key: "multi_signature_address", Value: k},
				{Key: "type", Value: multisig.MSTypeOwnerX},
			},
		}
		qs = append(qs, q)
	}
	resp, err := b.DBRead.searches(multisig.MSDBX, multisig.MSDBX, qs, decodeMSOwner)
	if err != nil {
		return err
	}
	for _, rs := range resp {
		for _, r := range rs {
			l, ok := r.(*multisig.MSOwner)
			if !ok {
				return errors.New(ErrTypeAsset)
			}
			outMS[l.MultiSigAddr].Owners = append(outMS[l.MultiSigAddr].Owners, l)
		}
	}
	return nil
}

// Searches Searches
func (b *MultiSig) Searches(reqs []*querypara.Query, out *interface{}) error {
	if len(reqs) == 0 {
		return nil
	}

	resp, err := b.DBRead.searches(multisig.MSDBX, multisig.MSDBX, reqs, decodeAllMS)
	if err != nil {
		return err
	}

	addrs, err := getAddresses(resp)
	if err != nil {
		return err
	}

	*out = addrs
	return err
}

// Search Search
func (b *MultiSig) Search(req *querypara.Query, out *interface{}) error {
	if req == nil {
		return nil
	}
	resp, err := b.DBRead.search(multisig.MSDBX, multisig.MSDBX, req, decodeAllMS)
	if err != nil {
		return err
	}

	addrs, err := getAddress(resp)
	if err != nil {
		return err
	}

	*out = addrs
	return err
}
