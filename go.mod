module github.com/arduino/mdns-discovery

go 1.25.2

replace github.com/hashicorp/mdns => github.com/cmaglie/mdns v1.0.6-0.20260615092725-1cef078b7fd6

require (
	github.com/arduino/go-properties-orderedmap v1.8.1
	github.com/arduino/pluggable-discovery-protocol-handler/v2 v2.2.1
	github.com/hashicorp/mdns v1.0.6
)

require (
	github.com/arduino/go-paths-helper v1.14.0 // indirect
	github.com/miekg/dns v1.1.72 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/tools v0.46.0 // indirect
)
