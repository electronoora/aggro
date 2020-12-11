/*
 * aggro
 *
 * Functions for outputting prefix lists in various formats
 *
 * Copyright (c) 2017 Noora Halme - please see LICENSE.md
 *
 */
package main

import (
  "fmt"
)



// output prefix list in Linux iptables format
func outputIptables(name string, prefixList *[]string) {
  fmt.Printf("iptables -F %s-ingress\n", name);
  fmt.Printf("iptables -F %s-egress\n", name);
  fmt.Printf("iptables -X %s-ingress\n", name);
  fmt.Printf("iptables -X %s-egress\n", name);
  fmt.Printf("iptables -N %s-ingress\n", name);
  fmt.Printf("iptables -N %s-egress\n", name);  
  for i := 0; i < len(*prefixList); i++ {
    fmt.Printf("iptables -A %s-ingress -s %s -j DROP\n", name, (*prefixList)[i]);
    fmt.Printf("iptables -A %s-egress -d %s -j DROP\n", name, (*prefixList)[i]);
  }
}


// output prefix list in Linux nptables format
func outputNptables(name string, prefixList *[]string) {
  fmt.Printf("nft add table ip %s\n", name);
  fmt.Printf("nft add chain ip %s input { type filter hook input priority 0 ; }\n", name);
  fmt.Printf("nft add chain ip %s output { type filter hook output priority 0 ; }\n", name);
  for i := 0; i < len(*prefixList); i++ {
    fmt.Printf("nft add rule ip %s input saddr %s drop\n", name, (*prefixList)[i]);
    fmt.Printf("nft add rule ip %s output daddr %s drop\n", name, (*prefixList)[i]);
  }  
}


// output prefix list in BSD/macOS pf format
func outputPf(name string, prefixList *[]string) {
  fmt.Printf("table <%s> {", name);
  for i := 0; i < len(*prefixList); i++ {
    fmt.Printf(" %s", (*prefixList)[i]);
  }
  fmt.Printf(" }\n");
  fmt.Printf("block drop in quick from { <%s> } to any\n", name);
  fmt.Printf("block drop out quick from any to { <%s> }\n", name);
}


// output prefix list in JunOS policy-options prefix-list
func outputJunos(name string, prefixList *[]string) {
  fmt.Printf("set policy-options prefix-list %s [ ", name);
  for i := 0; i < len(*prefixList); i++ {
    fmt.Printf("%s ", (*prefixList)[i]);
  }
  fmt.Printf("];\n");
  fmt.Printf("set firewall family inet filter %s term prefix-match from prefix-list %s then reject;\n", name, name);
  fmt.Printf("set firewall family inet filter %s term pass-through then accept;\n", name);
}


// output prefix list in plain CIDR format
func outputPlain(prefixList *[]string) {
  for i := 0; i < len(*prefixList); i++ {
    fmt.Printf("%s\n", (*prefixList)[i]);
  }
}
