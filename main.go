/*
 * aggro
 *
 * This is a utility for acquiring all the network prefixes announced by a list
 * of autonomous systems, aggregating prefixes into less specific ones and
 * removing prefixes that are already covered by less specific prefix. The final
 * list of prefixes should be the shortest list of prefixes that covers all the
 * address ranges advertised by the ASes.
 *
 * Additionally, prefixes can also be queried by association to a country specified
 * by an ISO-3166 alpha-2 country code.
 *
 * Please see README.md before using this utility.
 * 
 * This is my first Go project - please be gentle. :D This is more or less a
 * direct conversion from my node.js script into golang with some added features
 * and configurability.
 *
 * Copyright (c) 2017 Noora Halme - please see LICENSE.md
 */
package main

import (
  "flag"
  "fmt"
  "strconv"
  "log"
  "regexp"
)



var verbosityLevel int;
var outputFormat string;
var queryType string;

const OUTPUT_QUIET  = 0
const OUTPUT_NORMAL = 1
const OUTPUT_EXTRA  = 2
const OUTPUT_DEBUG  = 3



// convert from network prefix to trie index
func net2i(u uint32, p uint8) uint32 {
  return ((1<<p)-1)+(u>>(32-p))
}


// convert from trie index to network prefix
func i2net(u uint32, p uint8) uint32 {
  if p == 0 {
    return 0
  }
  return (((u+1)<<(32-p))&0xffffffff)
}


// convert from trie index to network prefix length
func i2plen(u uint32) uint32 {
  var p uint32
  for p = 1; p <= 32; p++ {
    if u<((1<<p)-1) {
      return p-1;
    }
  }
  return 0;
}


// format uint32 network prefix to dotted notation
func u32toipv4(u uint32) string {
  return fmt.Sprintf("%d.%d.%d.%d", ((u>>24)&0xff), ((u>>16)&0xff), ((u>>8)&0xff), (u&0xff))
}



///////////////////////////////////////////////////////////////////////////////
//
// main program
//
func main() {
  shortestPrefix := 32
  longestPrefix := 0
  trie := make([]uint8, 1<<25); // gaah
  ipv4Prefixes := make([]string, 0);
  ipv6Prefixes := make([]string, 0);
  asList := make([]int, 0);
  ccList := make([]string, 0);

  // read flags from command line
  flag.StringVar(&outputFormat, "o", "ipt", "Output format <ipt|nft|pf|junos|plain>");
  flag.IntVar(&verbosityLevel, "v", 1, "Output vebosity <0..3>");
  flag.StringVar(&queryType, "q", "as", "Query prefixes by ASNs or ISO country codes <as|cc>");
  flag.Parse()
  args := flag.Args()

  if len(args) < 1 {
    flag.PrintDefaults()
    return
  }

  if queryType == "as" {
    // read ASNs from command line
    for i:= 0; i < len(args); i++ {
      as, err := strconv.Atoi(args[i])
      if (err == nil) {
        asList=append(asList, as)
      }
    }

    // fetch network prefixes for all AS numbers specified on the command line
    asnText:=""
    for i := 0; i < len(asList); i++ {
      ann4, ann6, err := getASPrefixes(asList[i], &ipv4Prefixes, &ipv6Prefixes)
      if err != nil {
        log.Fatal("Failed to query prefixes for AS - aborting")
        return
      } else {
        asnText=fmt.Sprintf("%s%d ", asnText, asList[i])
        if verbosityLevel > OUTPUT_NORMAL {
          fmt.Printf("# AS%d announces %d IPv4 prefixes and %d IPv6 prefixes\n", asList[i], ann4, ann6)
        }
      }              
    }
    if verbosityLevel > OUTPUT_QUIET {
      fmt.Printf("# Successfully queried ASNs: %s\n", asnText);
    }
  } else {
    // args has ISO country codes
    ccText := ""
    for i:= 0; i < len(args); i++ {
      // cc is args[i]
      if len(args[i]) == 2 {
        ccList=append(ccList, args[i])
        // query the country from RIPEstat
        ann4, ann6, err := getCountryPrefixes(args[i], &ipv4Prefixes, &ipv6Prefixes)
        if err != nil {
          log.Fatal("Failed to query prefixes for country - aborting")
          return
        } else {
          ccText=fmt.Sprintf("%s%s ", ccText, args[i])
          if verbosityLevel > OUTPUT_NORMAL {
            fmt.Printf("# Country '%s' is associated with %d IPv4 prefixes and %d IPv6 prefixes\n", args[i], ann4, ann6)
          }
        }
      } else {
        // invalid country code
        log.Fatal("Invalid country code specified");
      }
    }
    if verbosityLevel > OUTPUT_QUIET {
      fmt.Printf("# Successfully queried CCs: %s\n", ccText);
    }    
  }

  // all network prefixes are now loaded - time to start aggregating them!
  if verbosityLevel > OUTPUT_QUIET {
    fmt.Printf("# Found a total of %d IPv4 prefixes and %d IPv6 prefixes\n", len(ipv4Prefixes), len(ipv6Prefixes));
  }
    
  // we'll aggregate prefixes by first arranging them into a trie. a node marked 0 means
  // no prefix, 1-32 means an exact or more specific prefix exists.
  // a node with an exact prefix is a leaf node in the tree.
  //
  // because generally no longer prefixes than /24 are accepted on the route table,
  // we can brute force it and allocate a 2^25 byte buffer to store our trie
  // breadth-first. storing as nodes with references would of course consume less
  // memory.

  for i := 0; i < len(ipv4Prefixes); i++ {
    prefixRE, err := regexp.Compile("^([0-9.]{7,15})/([0-9]{1,2})$")
    if err == nil {
      prefixMatches := prefixRE.FindStringSubmatch(ipv4Prefixes[i])
      if len(prefixMatches) == 3 {
        prefixLen, err := strconv.Atoi(prefixMatches[2])
        if err == nil {
          // skip longer prefixes than /24 - they wouldn't be routed anyway
          if prefixLen > 24 {
            if verbosityLevel > OUTPUT_NORMAL {
              fmt.Printf("# Skipping prefix %s for being longer than 24 bits\n", ipv4Prefixes[i])
            }
            continue 
          }

          // extract individual bytes from the prefix
          bytesRE, err := regexp.Compile("^([0-9]{1,3}).([0-9]{1,3}).([0-9]{1,3}).([0-9]{1,3})$")
          if err == nil {
            bytesMatches := bytesRE.FindStringSubmatch(prefixMatches[1])
            if len(bytesMatches) == 5 {
              // prefix looks ok, convert to uint32
              var network uint32
              for j := 0; j < 4; j++ {
                pb, err := strconv.Atoi(bytesMatches[1+j])
                if err == nil && pb >= 0 && pb < 256 {
                  network = network | (uint32(pb) << uint((3-j)*8))
                }
              }
              if verbosityLevel == OUTPUT_DEBUG {
                fmt.Printf("# Prefix %s (0x%08x len %d) has index 0x%08x\n", ipv4Prefixes[i], network, prefixLen, net2i(network, uint8(prefixLen)))
              }              
              
              // store the prefix into the trie
              for k :=1; k <= prefixLen; k++ {
                offset := net2i(network, uint8(k))
                if trie[offset] == 0 || trie[offset] > uint8(prefixLen) {
                  trie[offset]=uint8(prefixLen)
                }
              }              
            }
          } else {
            // regex failed to compile  
          }
        } else {
          log.Fatal("Invalid prefix length");
        }
               
        if prefixLen<shortestPrefix {
          shortestPrefix=prefixLen;
        }
        if prefixLen>longestPrefix {
          longestPrefix=prefixLen;
        }
      }
    } else {
      log.Fatal("Invalid IPv4 prefix");
    }
  }
  if verbosityLevel > OUTPUT_QUIET {
    fmt.Printf("# Shortest IPv4 prefix is %d bits and longest %d bits\n", shortestPrefix, longestPrefix);
  }

  // now traverse the trie, starting from the longest observed prefix length. if a node
  // and it's sibling are leaf nodes, they can be set to zero and the parent changed to a leaf.

  for p := longestPrefix; p >= shortestPrefix; p-- {
    for i := net2i(0, uint8(p)); i < net2i(0, uint8(p+1)); i+=2 {
      if trie[i]==uint8(p) && trie[i+1]==uint8(p) {
        if verbosityLevel > OUTPUT_NORMAL {
          fmt.Printf("# Aggregating %s/%d and %s/%d to %s/%d\n",
            u32toipv4(i2net(i,    uint8(p))), p,
            u32toipv4(i2net(i+1,  uint8(p))), p,
            u32toipv4(i2net(i>>1, uint8(p-1))), (p-1));
        }
        trie[i]=0;
        trie[i+1]=0;
        trie[i>>1]=uint8(p-1);
        if shortestPrefix > p-1 {
          shortestPrefix=p-1;
        }
      }
    }
    
    // scan through the prefixes again, this time to see if a less specific
    // prefix exists.

    for i := net2i(0, uint8(p)); i < net2i(0, uint8(p+1)); i++ {
      if trie[i] == uint8(p) {
      
        for pp := 1; pp <= (p-shortestPrefix); pp++ {
          pi := net2i(i2net(i, uint8(p)), uint8(p-pp));
          if trie[pi] == uint8(p-pp) {
            if verbosityLevel > OUTPUT_NORMAL {
              fmt.Printf("# Removing prefix %s/%d because less specific prefix %s/%d exists\n",
                u32toipv4(i2net(i,  uint8(p))), p,
                u32toipv4(i2net(pi, uint8(p-pp))), (p-pp));
            }              
            trie[i] = 0
            break;
          }
        }
      }
    }
  }

  // finally, traverse the trie once more time to collect all remaining leaf nodes. they are
  // used to generate the final list of prefixes

  aggregatedPrefixes := make([]string, 0)
  for p := longestPrefix; p >= shortestPrefix; p-- {
    for i := net2i(0, uint8(p)); i < net2i(0, uint8(p+1)); i++ {
      if trie[i] == uint8(p) {
        aggregatedPrefixes = append(aggregatedPrefixes, fmt.Sprintf("%s/%d", u32toipv4(i2net(i, uint8(p))), p))
      }
    }
  }
  if verbosityLevel > OUTPUT_QUIET {
    fmt.Printf("# Final list contains %d IPv4 prefixes\n", len(aggregatedPrefixes))
  }

  // now output the final list of prefixes collected from the trie
  listName := queryType; // "as" or "cc"
  switch queryType {
    case "as":
    for i := 0; i < len(asList); i++ {
      listName = fmt.Sprintf("%s-%d", listName, asList[i])
    }
    case "cc":
    for i := 0; i < len(ccList); i++ {
      listName = fmt.Sprintf("%s-%s", listName, ccList[i])
    }
  }
  switch outputFormat {
    case "ipt":
    outputIptables(listName, &aggregatedPrefixes)
    
    case "nft":
    outputNptables(listName, &aggregatedPrefixes)
    
    case "pf":
    outputPf(listName, &aggregatedPrefixes)
    
    case "junos":
    outputJunos(listName, &aggregatedPrefixes)
    
    case "plain":
    outputPlain(&aggregatedPrefixes)
  }
}
