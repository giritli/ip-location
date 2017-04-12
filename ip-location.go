package iplocation

import (
	"fmt"
	"net"
	"os"
	"bytes"
	"github.com/oschwald/geoip2-golang"
	"log"
)

const (
	CONN_HOST = ""
	CONN_TYPE = "tcp"
	ERR_INVALID_IP = 64
	ERR_COUNTRY_CODE_NOT_FOUND = 128
)

func Serve(file string, port string) {

	db, err := geoip2.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, ":" + port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + port)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		go handleRequest(conn, db)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn, db *geoip2.Reader) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.

	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		conn.Close()
		return
	}

	n := bytes.Index(buf, []byte{0})
	ip := net.ParseIP(string(buf[:n-1]))

	record, err := db.City(ip)

	if err != nil {
		conn.Write([]byte(fmt.Sprintf(`{"success":false,"errorCode":%d}`, ERR_INVALID_IP)))
		return
	}

	isoCode := record.Country.IsoCode
	json := "";

	if len(isoCode) == 2 {
		json = fmt.Sprintf(`{"success":true,"countryCode":"%s"}`, isoCode)
		fmt.Printf("%s -> %s\n", isoCode, ip)
	} else {
		json = fmt.Sprint(fmt.Sprintf(`{"success":false,"errorCode":%d}`, ERR_COUNTRY_CODE_NOT_FOUND))
	}

	conn.Write([]byte(json));
	conn.Close()
}