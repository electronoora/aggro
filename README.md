# aggro

This is a utility for acquiring all the network prefixes announced by a list of autonomous systems, aggregating prefixes into less specific ones and removing prefixes that are already covered by another prefix. The resulting list of prefixes should be the shortest list of prefixes that cover all the address ranges advertised by the ASNs.

The code uses [RIPEstat data API](https://stat.ripe.net/docs/data_api) to fetch lists of announced prefixes per AS. Please let them know if you run large lists of ASNs or do over 1000 requests per day.

Additionally, prefixes can also be queried by association to a country specified by an ISO-3166 alpha-2 country code.

Output can either output a plain list of prefixes, Linux iptables or nft rules, BSD pf rules or JunOS firewall filter statements. Note that the output types have not been tested yet in actual use. :D

My initial reason for writing this was a desire to have an easy way to automatically generate firewall rules that apply to a particular list of service providers and keep the total number of rules as low as possible.

Two other reasons were that I wanted to learn Go and experiment with a trie (aka. radix tree) for storing and searching IPv4 routes. :)

For information, explanations and diagrams on how tries are used with IPv4 prefixes, check out these pages:

https://www.juniper.net/documentation/en_US/junos/topics/usage-guidelines/policy-configuring-route-lists-for-use-in-routing-policy-match-conditions.html

https://vincent.bernat.im/en/blog/2017-ipv4-route-lookup-linux



### Command line arguments

The utility accepts the following command line arguments:

```
Usage of ./aggro:
  -o string
        Output format <plain|ipt|nft|pf|junos> (default "ipt")
  -q string
        Query prefixes by ASNs or ISO country codes <as|cc> (default "as")
  -v int
        Output vebosity <0..3> (default 1)
```



### Examples

```
$ ./aggro -o ipt 32934 54115 63293
# Successfully queried ASNs: 32934 54115 63293 
# Found a total of 77 IPv4 prefixes and 144 IPv6 prefixes
# Shortest prefix is 17 bits and longest 24 bits
# Final list contains 15 IPv4 prefixes
iptables -F as-32934-54115-63293-ingress
iptables -F as-32934-54115-63293-egress
iptables -X as-32934-54115-63293-ingress
iptables -X as-32934-54115-63293-egress
iptables -N as-32934-54115-63293-ingress
iptables -N as-32934-54115-63293-egress
iptables -A as-32934-54115-63293-ingress -s 157.240.163.0/24 -j DROP
iptables -A as-32934-54115-63293-egress -d 157.240.163.0/24 -j DROP
iptables -A as-32934-54115-63293-ingress -s 45.64.40.0/22 -j DROP
iptables -A as-32934-54115-63293-egress -d 45.64.40.0/22 -j DROP
iptables -A as-32934-54115-63293-ingress -s 74.119.76.0/22 -j DROP
iptables -A as-32934-54115-63293-egress -d 74.119.76.0/22 -j DROP
iptables -A as-32934-54115-63293-ingress -s 103.4.96.0/22 -j DROP
iptables -A as-32934-54115-63293-egress -d 103.4.96.0/22 -j DROP
iptables -A as-32934-54115-63293-ingress -s 179.60.192.0/22 -j DROP
iptables -A as-32934-54115-63293-egress -d 179.60.192.0/22 -j DROP
iptables -A as-32934-54115-63293-ingress -s 185.60.216.0/22 -j DROP
iptables -A as-32934-54115-63293-egress -d 185.60.216.0/22 -j DROP
iptables -A as-32934-54115-63293-ingress -s 199.201.64.0/22 -j DROP
iptables -A as-32934-54115-63293-egress -d 199.201.64.0/22 -j DROP
iptables -A as-32934-54115-63293-ingress -s 204.15.20.0/22 -j DROP
iptables -A as-32934-54115-63293-egress -d 204.15.20.0/22 -j DROP
iptables -A as-32934-54115-63293-ingress -s 31.13.24.0/21 -j DROP
iptables -A as-32934-54115-63293-egress -d 31.13.24.0/21 -j DROP
iptables -A as-32934-54115-63293-ingress -s 66.220.144.0/20 -j DROP
iptables -A as-32934-54115-63293-egress -d 66.220.144.0/20 -j DROP
iptables -A as-32934-54115-63293-ingress -s 69.63.176.0/20 -j DROP
iptables -A as-32934-54115-63293-egress -d 69.63.176.0/20 -j DROP
iptables -A as-32934-54115-63293-ingress -s 69.171.224.0/19 -j DROP
iptables -A as-32934-54115-63293-egress -d 69.171.224.0/19 -j DROP
iptables -A as-32934-54115-63293-ingress -s 31.13.64.0/18 -j DROP
iptables -A as-32934-54115-63293-egress -d 31.13.64.0/18 -j DROP
iptables -A as-32934-54115-63293-ingress -s 173.252.64.0/18 -j DROP
iptables -A as-32934-54115-63293-egress -d 173.252.64.0/18 -j DROP
iptables -A as-32934-54115-63293-ingress -s 157.240.0.0/17 -j DROP
iptables -A as-32934-54115-63293-egress -d 157.240.0.0/17 -j DROP
```

```
$ ./aggro -o junos -q cc sc
# Successfully queried CCs: sc
# Found a total of 66 IPv4 prefixes and 12 IPv6 prefixes
# Shortest IPv4 prefix is 11 bits and longest 24 bits
# Final list contains 56 IPv4 prefixes
<output removed>
```



### Compiling

Make sure first that you have the Go compiler and utilities installed on your
system:

```
$ go version
go version go1.9.2 darwin/amd64
```

If everything looks good, find a directory where you want to clone the
source code to and type:

```
$ git clone https://github.com/electronoora/aggro.git
$ cd aggro
$ go build
```

