// SPDX-FileCopyrightText: (c) Mauve Mailorder Software GmbH & Co. KG, 2022. Licensed under [MIT](LICENSE) license.
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// parseTarget validates the host parameter of a scrape request and returns it in a form
// usable as the host part of an URL. Anything but host[:port] is rejected, so a crafted
// parameter can not redirect the request (and the credentials) to another path or host.
func parseTarget(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("no host defined")
	}

	invalid := fmt.Errorf("invalid host %q", raw)
	if !utf8.ValidString(raw) || strings.ContainsAny(raw, "/?#@\\%") {
		return "", invalid
	}
	if strings.ContainsFunc(raw, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
		return "", invalid
	}

	addr := raw
	if ip, err := netip.ParseAddr(raw); err == nil && ip.Is6() {
		addr = "[" + raw + "]"
	}

	u, err := url.Parse("https://" + addr)
	if err != nil || u.Host != addr || u.Hostname() == "" {
		return "", invalid
	}

	if p := u.Port(); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 65535 {
			return "", invalid
		}
	}

	return addr, nil
}

// targetFilter restricts the hosts which can be scraped
type targetFilter struct {
	hosts    map[string]bool
	prefixes []netip.Prefix
}

// parseTargetFilter parses a comma separated list of host names, IP addresses and CIDR prefixes.
// An empty list allows all targets.
func parseTargetFilter(s string) (*targetFilter, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}

	f := &targetFilter{hosts: make(map[string]bool)}
	for _, e := range strings.Split(s, ",") {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}

		if strings.Contains(e, "/") {
			p, err := netip.ParsePrefix(e)
			if err != nil {
				return nil, fmt.Errorf("invalid allowed target %q: %w", e, err)
			}
			f.prefixes = append(f.prefixes, p.Masked())
			continue
		}

		if ip, err := netip.ParseAddr(e); err == nil {
			f.prefixes = append(f.prefixes, netip.PrefixFrom(ip.Unmap(), ip.Unmap().BitLen()))
			continue
		}

		f.hosts[strings.ToLower(e)] = true
	}

	return f, nil
}

// allows checks if the target (as returned by parseTarget) may be scraped.
// IP prefixes only match targets given as IP address, host names are not resolved.
func (f *targetFilter) allows(addr string) bool {
	if f == nil {
		return true
	}

	host := addr
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")

	if f.hosts[strings.ToLower(host)] {
		return true
	}

	ip, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}

	for _, p := range f.prefixes {
		if p.Contains(ip.Unmap()) {
			return true
		}
	}

	return false
}
