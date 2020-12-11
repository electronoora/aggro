/*
 * aggro
 *
 * Functions for querying data from the RIPEstat API
 *
 * Copyright (c) 2017 Noora Halme - please see LICENSE.md
 *
 */
package main

import (
  "errors"
  "fmt"
  "net/http"  
  "io/ioutil"
  "strings"
  "encoding/json"
)



// fetch a list of prefixes associated with a country matching an ISO-3166 alpha-2 country code
// append the slices provided as arguments with the found prefixes
func getCountryPrefixes(cc string, ipv4Slice *[]string, ipv6Slice *[]string) (int, int, error) {
  ann4 := 0; ann6 := 0
  url := fmt.Sprintf("https://stat.ripe.net/data/country-resource-list/data.json?resource=%s&v4_format=prefix", cc)
  res, err := http.Get(url);
  if err == nil {
    bytes, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err == nil {
      var data map[string]interface{}
      if err := json.Unmarshal(bytes, &data); err != nil {
        err := errors.New("JSON parsing failed")
        return 0, 0, err
      }
      if data["status"] == "ok" {
        // got the list of resources - read the ipv4 and ipv6 prefixes
        var prefixes []interface{}
        prefixes = data["data"].(map[string]interface{})["resources"].(map[string]interface{})["ipv4"].([]interface{})
        for j := 0; j < len(prefixes); j++ { 
          *ipv4Slice=append(*ipv4Slice, prefixes[j].(string));
          ann4++
        }
        prefixes = data["data"].(map[string]interface{})["resources"].(map[string]interface{})["ipv6"].([]interface{})
        for j := 0; j < len(prefixes); j++ { 
          *ipv6Slice=append(*ipv6Slice, prefixes[j].(string));
          ann6++
        }
      }
    } else {
      return 0, 0, errors.New("Reading document body failed")
    }
  } else {
    return 0, 0, errors.New("HTTP request failed")
  }
  return ann4, ann6, nil
}


// fetch a list of all announced prefixes for an AS using the RIPEstat API
// append the slices provided as arguments with the found prefixes
func getASPrefixes(as int, ipv4Slice *[]string, ipv6Slice *[]string) (int, int, error) {
  ann4 := 0; ann6 := 0
  url := fmt.Sprintf("https://stat.ripe.net//data/announced-prefixes/data.json?resource=AS%d", as);
  res, err := http.Get(url);
  if err == nil {
    bytes, err := ioutil.ReadAll(res.Body)
    res.Body.Close()
    if err == nil {
      var data map[string]interface{}
      if err := json.Unmarshal(bytes, &data); err != nil {
        err := errors.New("JSON parsing failed")
        return 0, 0, err
      }
      if data["status"] == "ok" {
        prefixes := data["data"].(map[string]interface{})["prefixes"].([]interface{})
        for j := 0; j < len(prefixes); j++ {
          prefix := prefixes[j].(map[string]interface{})["prefix"].(string)
          if strings.ContainsRune(prefix, ':') {
            //fmt.Printf("# IPv6: %s\n", prefix)
            *ipv6Slice=append(*ipv6Slice, prefix);
            ann6++
          } else {
            //fmt.Printf("# IPv4: %s\n", prefix)
            *ipv4Slice=append(*ipv4Slice, prefix);
            ann4++
          }
        }
      }
    } else {
      return 0, 0, errors.New("Reading document body failed")
    }
  } else {
    return 0, 0, errors.New("HTTP request failed")
  }
  return ann4, ann6, nil
}
