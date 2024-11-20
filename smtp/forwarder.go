package server

import (
	"errors"
	"net"
)

func lookupMx(domain string) ([]*net.MX, error) {
	record, err := net.LookupMX(domain)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func sendMailToMx(records []*net.MX, mail Mail) error {
	for _, record := range records {
		addr := record.Host + ":25"
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			continue
		}

		defer conn.Close()

		// TODO: sending mail here

		return nil
	}

	return errors.New("could not send mail to any mx")
}
