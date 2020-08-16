# CLI_Ping_Application

A small Ping CLI application accepts hostname or IP address (IPV4 and IPV6) through rootCmd and sends ICMP echo requests while receiving echo replies. Written using Go and Cobra. Reports loss and RTT times, and can set TTL as argument. Must use root privileges. If only state DNS/IP address, then TTL automatically set to 10 seconds.

Ex for running on Linux: "sudo ./my-ping google.com 5" (TTL set for 5 seconds)

## Run/Installation:
sudo go install my-ping

sudo go build
