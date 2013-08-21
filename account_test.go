package golink

import (
"time"
"testing"
"net/url"
"code.google.com/p/go-etree"
"bytes"
)

type apiTester string

func (r apiTester) Get(path string, params url.Values, c *APICredentials) (t etree.Element, err error) {
  t, err = etree.Parse(bytes.NewBufferString(string(r)))
  if err != nil {
    return t, err
  }
  return t.Item(0), nil
}

func TestStatus(t *testing.T) {
  a := NewCredentialedAPI(apiTester(statusXML), APICredentials{})
  status, err := a.AccountStatus()
  if err != nil {
    t.Error(err)
  }
  if status != (aStatus{Paid: time.Unix(1293840000, 0).UTC(), Created: time.Unix(1072915200, 0).UTC(), Logins: 1234, MinutesPlayed: 9999}) {
    t.Errorf("Wrong status returned. Got %+v", status)
  }
}

func TestKeyInfo(t *testing.T) {
  a := NewCredentialedAPI(apiTester(keyInfoXML), APICredentials{})
  kinfo, err := a.AccountKeyInfo()
  if err != nil {
    t.Error(err)
  }
  x := time.Unix(1315699200, 0).UTC()
  if kinfo.AccessMask != 59760264 || kinfo.Type != "char" || *kinfo.Expires != x || len(kinfo.Characters) != 1 || kinfo.Characters[898901870] != (aCharacter{Id: 898901870, Name: "Desmont McCallock", CorpId: 1000009, CorpName: "Caldari Provisions",}) {
    t.Errorf("Wrong key info returned. Got %+v", kinfo)
  }
}

func TestCharacters(t *testing.T) {
  a := NewCredentialedAPI(apiTester(charactersXML), APICredentials{})
  chars, err := a.AccountCharacters()
  if err != nil {
    t.Error(err)
  }
  if chars[1365215823] != (aCharacter{Id: 1365215823, Name: "Alexis Prey", CorpId: 238510404, CorpName: "Puppies To the Rescue"}) {
    t.Errorf("Wrong character info returned. Got %+v", chars)
  }
}

const (
  statusXML = `
<result>
    <paidUntil>2011-01-01 00:00:00</paidUntil>
    <createDate>2004-01-01 00:00:00</createDate>
    <logonCount>1234</logonCount>
    <logonMinutes>9999</logonMinutes>
</result>
`
  charactersXML = `
<result>
    <rowset name="characters">
        <row name="Alexis Prey" characterID="1365215823"
         corporationName="Puppies To the Rescue" corporationID="238510404"/>
    </rowset>
</result>
`
  keyInfoXML = `
<result>
    <key accessMask="59760264" type="Character" expires="2011-09-11 00:00:00">
        <rowset name="characters">
            <row characterID="898901870" characterName="Desmont McCallock"
             corporationID="1000009" corporationName="Caldari Provisions" />
        </rowset>
    </key>
</result>
`
)
