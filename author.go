/*
	fourcolor - a simple anonymous bbs
	Copyright (C) 2025  Pancakes <patapancakes@pagefault.games>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"os"
)

// AUTHOR
var authorKey []byte

func getAuthorKey() ([]byte, error) {
	key, err := os.ReadFile("author.key")
	if err == nil {
		return key, nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	buf := make([]byte, 256)
	_, err = rand.Read(buf)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile("author.key", buf, 0600)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func deriveAuthorDigest(ip string, id int) []byte {
	hash := sha256.New()

	// add author key
	hash.Write(authorKey)

	// add ip
	hash.Write([]byte(ip))

	// add post id
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(id))
	hash.Write(buf)

	return hash.Sum(nil)
}

// TICKET
var ticketKey []byte

func getTicketKey() ([]byte, error) {
	key, err := os.ReadFile("ticket.key")
	if err == nil {
		return key, nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	buf := make([]byte, 256)
	_, err = rand.Read(buf)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile("ticket.key", buf, 0600)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func deriveTicket(ip string, ua string) []byte {
	hash := sha256.New()

	// add ticket key
	hash.Write(ticketKey)

	// add ip
	hash.Write([]byte(ip))

	// add user agent
	hash.Write([]byte(ua))

	return hash.Sum(nil)
}
