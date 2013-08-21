package golink

import(
"fmt"
"time"
"strconv"
"strings"
  "code.google.com/p/go-etree"
)

func parseEveTs(in string) (time.Time, error) {
  return time.Parse("2006-01-02 15:04:05", in)
}

func first(s string, _ bool) string {
  return s
}

func getStrValue(e etree.Element, child string) (string, error) {
  c := e.Find(child)
  if c == nil {
    return "", fmt.Errorf("Unable to find child element %v of element %v", child, e.Tag())
  }
  return c.Text(), nil
}

func getTimeValue(e etree.Element, child string) (time.Time, error) {
  c := e.Find(child)
  if c == nil {
    return time.Time{}, fmt.Errorf("Unable to find child element %v of element %v", child, e.Tag())
  }
  return parseEveTs(c.Text())
}

func getIntValue(e etree.Element, child string) (int64, error) {
  c := e.Find(child)
  if c == nil {
    return 0, fmt.Errorf("Unable to find child element %v of element %v", child, e.Tag())
  }
  return strconv.ParseInt(c.Text(), 0, 64)
}

func getFloatValue(e etree.Element, child string) (float64, error) {
  c := e.Find(child)
  if c == nil {
    return 0, fmt.Errorf("Unable to find child element %v of element %v", child, e.Tag())
  }
  return strconv.ParseFloat(c.Text(), 64)
}

func getBoolValue(e etree.Element, child string) (bool, error) {
  c := e.Find(child)
  if c == nil {
    return false, fmt.Errorf("Unable to find child element %v of element %v", child, e.Tag())
  }
  if c.Text() == "True" {
    return true, nil
  } else if c.Text() == "False" {
    return false, nil
  }
  return false, fmt.Errorf("Unknown bool literal value: %v", c.Text())
}

func extractKeyval(data string) (map[string]string) {
  pairs := strings.FieldsFunc(data, func(r rune) bool {return r == '\n'})
  ret := make(map[string]string)
  for _, s := range pairs {
    r := strings.Split(s, ": ")
    ret[r[0]] = r[1]
  }
  return ret
}

func parseMSDate(data string) (time.Time, error) {
  i, err := strconv.ParseInt(data, 0, 64)
  if err != nil {
    return time.Time{}, err
  }
  i = (i/10000000) - 11644473600
  return time.Unix(i, 0), nil
}
