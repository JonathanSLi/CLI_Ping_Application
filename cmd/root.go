/*
2020 Jonathan Li <jonathan.li.ttexpert@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"net"
	"strconv"
    "time"
    "log"
    "golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var ListenAddr = "0.0.0.0"

var (
	packetsReceived int = 0
	packetsSent int = 0
	loss float64 = 0
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "my-ping",
	Short: "A small Ping CLI application accepts hostname or IP address \n as rootCmd and sends ICMP echo requests while receiving echo replies",
	Long: 
	`A small Ping CLI application accepts hostname or IP address (IPV4 and IPV6)
	through rootCmd and sends ICMP echo requests while receiving echo replies. Written 
	using Go and Cobra. Reports loss and RTT times, set TTL as argument. Must use root
	privileges. If only state DNS/IP address, then TTL automatically set for 10 seconds
	Ex for running on Linux: "sudo ./my-ping google.com 5" (TTL set for 5 seconds)
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 2 {
			log.Printf("Needs 1 or 2 argument: IP Address/DNS and/or TTL")
			return
		}
		addr := args[0]
		var TTL int
		if len(args) == 1 {
			TTL = 10
		}else{
			TTL1, err := strconv.Atoi(args[1])
			TTL = TTL1
			if err != nil {
				log.Printf("Incorrect second argument, not an integer")
				return
			}
		}
		for {
			Ping(addr, TTL)
			time.Sleep(1 * time.Second)
		}
		
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.my-ping.yaml)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".my-ping" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".my-ping")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func Helper_Ping(addr string, TTL int) (*net.IPAddr, time.Duration, float64, error) {
	
	dst, err := net.ResolveIPAddr("ip4", addr)
	dst1, err1:= net.ResolveIPAddr("ip6", addr)
	var isip6 bool = false
    if err != nil {
		
		//if both ipv4 and iv6 do not work
		if err1 != nil{
			return nil, 0, 0, err
		}else{
			log.Printf("ipv6")
			isip6 = true
			err = err1
			dst = dst1
		}
	}

	protocol := 1
	var type_ icmp.Type
	network := "ip4:icmp"
	type_ = ipv4.ICMPTypeEcho
	if isip6 {
		protocol = 58
		network = "ip6:icmp"
		type_ = ipv6.ICMPTypeEchoRequest
	}
	

    //listen for icmp replies
    c, err := icmp.ListenPacket(network, ListenAddr)
    if err != nil {
        return nil, 0, 0, err
    }
    defer c.Close()

	// Make a ICMP message catching necessary error messages.
	
	m := icmp.Message{
		Type: type_, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte(""),
		},
	}
	
    b, err := m.Marshal(nil)
    if err != nil {
        return dst, 0, 0, err
    }
	// Sending request
    start := time.Now()
    n, err := c.WriteTo(b, dst)
    if err != nil {
        return dst, 0, 0, err
    } else if n != len(b) {
        return dst, 0, 0, fmt.Errorf("got %v; want %v", n, len(b))
	}
	packetsSent += 1
	//log.Printf("%v", packetsSent)

    // Wait for a reply
	reply := make([]byte, 1500)
    err = c.SetReadDeadline(time.Now().Add(time.Duration(TTL) * time.Second))
    if err != nil {
		loss = float64(packetsSent-packetsReceived) / float64(packetsSent) * 100
		
        return dst, 0, loss, err
    }
    n, peer, err := c.ReadFrom(reply)
    if err != nil {
		loss = float64(packetsSent-packetsReceived) / float64(packetsSent) * 100
        return dst, 0, loss, err
    }
    duration := time.Since(start)
    rm, err := icmp.ParseMessage(protocol, reply[:n])
    if err != nil {
        return dst, 0, 0, err
	}

    switch rm.Type {
	case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
		packetsReceived += 1

		loss = float64(packetsSent-packetsReceived) / float64(packetsSent) * 100
        return dst, duration, loss, nil
	default:
		loss = float64(packetsSent-packetsReceived) / float64(packetsSent) * 100
        return dst, 0, loss, fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
    }
}

func Ping(addr string, TTL int) {
	
    dst, dur, loss, err := Helper_Ping(addr, TTL)
    if err != nil {
        log.Printf("Ping %s (%s): Loss: %g percent Time: %s\n", addr, dst, loss, err)
        return
    }
    log.Printf("Ping %s (%s): Loss: %g percent Time: %s\n", addr, dst, loss, dur)
} 