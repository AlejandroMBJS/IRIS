package nfc

import (
	"errors"
	"fmt"
	"github.com/ebfe/scard"
	"strings"
)

type Reader struct {
	ctx        *scard.Context
	readerName string
}

func NewReader() (*Reader, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, fmt.Errorf("Eror %v", err)
	}
	readers, err := ctx.ListReaders()
	if err != nil {
		return nil, fmt.Errorf("No se encontraron lectores conectados: %v", err)
	}
	var readerName string
	for _, r := range readers {
		if strings.Contains(r, "ACR122") {
			readerName = r
			break
		}
	}
	if readerName == "" {
		return nil, errors.New("ACR122U no encontrado")
	}
	return &Reader{ctx: ctx, readerName: readerName}, nil
}

func (r *Reader) WaitForCard() (string, error) {
	card, err := r.ctx.Connect(r.readerName, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		return "", fmt.Errorf("Error conectando: %v", err)
	}
	defer card.Disconnect(scard.LeaveCard)

	getUID := []byte{0xFF, 0xCA, 0x00, 0x00, 0x00}

	response, err := card.Transmit(getUID)
	if err != nil {
		return "", fmt.Errorf("Error leyendo la tarjeta: %v", err)
	}

	if len(response) < 3 {
		return "", errors.New("Respuesta Invalida")
	}

	sw1 := response[len(response)-2]
	sw2 := response[len(response)-1]
	if sw1 != 0x90 || sw2 != 0x00 {
		return "", errors.New("Tarjeta rechazo comando")
	}

	uid := response[:len(response)-2]
	uidhex := strings.ToUpper(fmt.Sprintf("%x", uid))
	return uidhex, nil
}

func (r *Reader) Close() error {
	if r.ctx != nil {
		return r.ctx.Release()
	}
	return nil
}
