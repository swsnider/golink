package parser

import (
  "bytes"
  "encoding/xml"
  "reflect"
)

func Unmarshal(data []byte, v interface{}) error {
  r := bytes.NewReader(data)
  d := xml.NewDecoder(r)
  for t, err = d.Token(); t != nil; t, err = d.Token() {
    switch token := t.(type) {
    case xml.StartElement:
      return nil
    case xml.EndElement:
      return nil
    case xml.CharData:
      return nil
    case xml.Directive:
      return nil
    case xml.ProcInst:
      continue // We don't care.
    case xml.Comment:
      continue // Certainly don't care.
    }
  }
  return nil
}
