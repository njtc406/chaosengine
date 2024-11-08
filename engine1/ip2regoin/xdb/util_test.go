// Copyright 2022 The Ip2Region Authors. All rights reserved.
// Use of this source code is governed by a Apache2.0-style
// license that can be found in the LICENSE file.

// ---
// @Author Lion <chenxin619315@gmail.com>
// @Date   2022/06/16

package xdb

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestCheckIP(t *testing.T) {
	var str = "29.34.191.255"
	ip, err := CheckIP(str)
	if err != nil {
		t.Errorf("check ip `%s`: %s\n", str, err)
	}

	netIP := net.ParseIP(str).To4()
	if netIP == nil {
		t.Fatalf("parse ip `%s` failed", str)
	}

	u32 := binary.BigEndian.Uint32(netIP)
	fmt.Printf("checkip: %d, parseip: %d, isEqual: %v\n", ip, u32, ip == u32)
}

func TestLong2IP(t *testing.T) {
	var str = "29.34.191.255"
	netIP := net.ParseIP(str).To4()
	if netIP == nil {
		t.Fatalf("parse ip `%s` failed", str)
	}

	u32 := binary.BigEndian.Uint32(netIP)
	ipStr := Long2IP(u32)
	fmt.Printf("originIP: %s, Long2IP: %s, isEqual: %v\n", str, ipStr, ipStr == str)
}

func TestLoadVectorIndex(t *testing.T) {
	// 1、从 dbPath 加载整个 xdb 到内存
	cBuff, err := LoadContentFromFile("../ip2region.xdb")
	if err != nil {
		fmt.Printf("failed to load content from `%s`: %s\n", "../ip2region.xdb", err)
		return
	}

	// 2、用全局的 cBuff 创建完全基于内存的查询对象。
	XSearcher, err := NewWithBuffer(cBuff)
	if err != nil {
		fmt.Printf("failed to create searcher with content: %s\n", err)
		return
	}

	region, err := XSearcher.SearchByStr("29.34.191.255")
	fmt.Printf("region %s errdef %v", region, err)
}

func TestLoadContent(t *testing.T) {
	buff, err := LoadContentFromFile("../ip2region.xdb")
	if err != nil {
		fmt.Printf("failed to load xdb content: %s\n", err)
		return
	}

	fmt.Printf("buff length: %d\n", len(buff))
}

func TestLoadHeader(t *testing.T) {
	header, err := LoadHeaderFromFile("../ip2region.xdb")
	if err != nil {
		fmt.Printf("failed to load xdb header info: %s\n", err)
		return
	}

	fmt.Printf("Version        : %d\n", header.Version)
	fmt.Printf("IndexPolicy    : %s\n", header.IndexPolicy.String())
	fmt.Printf("CreatedAt      : %d(%s)\n", header.CreatedAt, time.Unix(int64(header.CreatedAt), 0).Format(time.RFC3339))
	fmt.Printf("StartIndexPtr  : %d\n", header.StartIndexPtr)
	fmt.Printf("EndIndexPtr    : %d\n", header.EndIndexPtr)
}
