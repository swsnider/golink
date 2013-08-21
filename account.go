package golink

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type aStatus struct {
	Paid          time.Time
	Created       time.Time
	Logins        int64
	MinutesPlayed int64
}

var key_types = map[string]string{"Account": "account", "Character": "char", "Corporation": "corp"}

func (a *CredentialedAPI) AccountStatus() (aStatus, error) {
	result, err := a.Get("account/AccountStatus", url.Values{})
	r := aStatus{}
	if err != nil {
		return r, err
	}
	if r.Paid, err = getTimeValue(result, "paidUntil"); err != nil {
		return r, err
	}
	if r.Created, err = getTimeValue(result, "createDate"); err != nil {
		return r, err
	}
	if r.Logins, err = getIntValue(result, "logonCount"); err != nil {
		return r, err
	}
	if r.MinutesPlayed, err = getIntValue(result, "logonMinutes"); err != nil {
		return r, err
	}
	return r, nil
}

type aCharacter struct {
	Id       int64
	Name     string
	CorpId   int64
	CorpName string
}

type aKInfo struct {
	AccessMask int64
	Type       string
	Expires    *time.Time
	Characters map[int64]aCharacter
}

func (a *CredentialedAPI) AccountKeyInfo() (aKInfo, error) {
	result, err := a.Get("account/APIKeyInfo", url.Values{})
	var ok bool
	r := aKInfo{}
	if err != nil {
		return r, err
	}
	result = result.Find("key")
	if r.AccessMask, err = strconv.ParseInt(first(result.Get("accessMask")), 0, 64); err != nil {
		return r, err
	}
	if r.Type, ok = key_types[first(result.Get("type"))]; ok == false {
		return r, fmt.Errorf("No key type returned from API server.")
	}
	if temp, ok := result.Get("expires"); ok == false {
		r.Expires = nil
	} else {
		temp, err := parseEveTs(temp)
		if err != nil {
			return r, err
		}
		r.Expires = &temp
	}

	r.Characters = make(map[int64]aCharacter)

	rowset := result.Find("rowset")
	for _, row := range rowset.FindAll("row") {
		c := aCharacter{}
		if c.Id, err = strconv.ParseInt(first(row.Get("characterID")), 0, 64); err != nil {
			return r, err
		}
		if c.Name, ok = row.Get("characterName"); ok == false {
			return r, fmt.Errorf("Unable to extract a character name.")
		}
		if c.CorpId, err = strconv.ParseInt(first(row.Get("corporationID")), 0, 64); err != nil {
			return r, err
		}
		if c.CorpName, ok = row.Get("corporationName"); ok == false {
			return r, fmt.Errorf("Unable to extract a corp name for %v.", c.Name)
		}
		r.Characters[c.Id] = c
	}
	return r, nil
}

func (a *CredentialedAPI) AccountCharacters() (map[int64]aCharacter, error) {
	result, err := a.Get("account/Characters", url.Values{})
	r := make(map[int64]aCharacter)
	var ok bool
	if err != nil {
		return nil, err
	}
	rowset := result.Find("rowset")
	if rowset == nil {
		return nil, fmt.Errorf("Unable to extract rowset from API response.")
	}
	for _, row := range rowset.FindAll("row") {
		c := aCharacter{}
		if c.Id, err = strconv.ParseInt(first(row.Get("characterID")), 0, 64); err != nil {
			return nil, err
		}
		if c.Name, ok = row.Get("name"); ok == false {
			return nil, fmt.Errorf("Unable to extract a character name.")
		}
		if c.CorpId, err = strconv.ParseInt(first(row.Get("corporationID")), 0, 64); err != nil {
			return nil, err
		}
		if c.CorpName, ok = row.Get("corporationName"); ok == false {
			return nil, fmt.Errorf("Unable to extract a corp name for %v.", c.Name)
		}
		r[c.Id] = c
	}
	return r, nil
}
