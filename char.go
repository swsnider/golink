package golink

import (
	//	"fmt"
	"code.google.com/p/go-etree"
	"net/url"
	"strconv"
	"time"
)

type cAsset struct {
	Id, ItemTypeId, Quantity, LocationId, LocationFlag int64
	Packaged                                           bool
	Contents                                           []cAsset
}

func handleAssetRowset(rowset etree.Element, locId int64) ([]cAsset, error) {
	ret := make([]cAsset, 0)
	var err error
	for _, row := range rowset.FindAll("row") {
		asset := cAsset{}
		if asset.Id, err = strconv.ParseInt(first(row.Get("itemID")), 0, 64); err != nil {
			return ret, err
		}
		if asset.ItemTypeId, err = strconv.ParseInt(first(row.Get("typeID")), 0, 64); err != nil {
			return ret, err
		}
		if asset.LocationId, err = strconv.ParseInt(first(row.Get("locationID")), 0, 64); err != nil {
			return ret, err
		}
		if asset.LocationFlag, err = strconv.ParseInt(first(row.Get("flag")), 0, 64); err != nil {
			return ret, err
		}
		if asset.Quantity, err = strconv.ParseInt(first(row.Get("quantity")), 0, 64); err != nil {
			return ret, err
		}
		asset.Packaged = first(row.Get("singleton")) == "0"
		contents := row.Find("rowset")
		if contents == nil {
			asset.Contents, err = handleAssetRowset(contents, asset.LocationId)
			if err != nil {
				return ret, err
			}
		}
		ret = append(ret, asset)
	}
	return ret, nil
}

func (a *CredentialedAPI) CharAssets(charId int64) ([]cAsset, error) {
	result, err := a.Get("char/AssetList", url.Values{"characterID": []string{string(charId)}})
	if err != nil {
		return nil, err
	}
	rowset := result.Find("rowset")
	if rowset == nil {
		return nil, err
	}
	return handleAssetRowset(rowset, -1)
}

type cContractBid struct {
	Id, ContractId, BidderId, Amount int64
	Timestamp                        time.Time
}

func (a *CredentialedAPI) CharContractBids(charId int64) ([]cContractBid, error) {
	result, err := a.Get("char/ContractBids", url.Values{"characterID": []string{string(charId)}})
	if err != nil {
		return nil, err
	}
	rowset := result.Find("rowset")
	if rowset == nil {
		return nil, err
	}
	r := make([]cContractBid, 0)
	for _, row := range rowset.FindAll("row") {
		c := cContractBid{}
		if c.Id, err = strconv.ParseInt(first(row.Get("bidID")), 0, 64); err != nil {
			return nil, err
		}
		if c.ContractId, err = strconv.ParseInt(first(row.Get("contractID")), 0, 64); err != nil {
			return nil, err
		}
		if c.BidderId, err = strconv.ParseInt(first(row.Get("bidderID")), 0, 64); err != nil {
			return nil, err
		}
		if c.Amount, err = strconv.ParseInt(first(row.Get("amount")), 0, 64); err != nil {
			return nil, err
		}
		if c.Timestamp, err = parseEveTs(first(row.Get("dateBid"))); err != nil {
			return nil, err
		}
		r = append(r, c)
	}
	return r, nil
}
