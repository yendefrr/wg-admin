package model

import (
	"bytes"
	"fmt"
	"io"
	"os"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/skip2/go-qrcode"
)

type Profile struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Type       string `json:"type"`
	Path       string `json:"path"`
	Publickey  string `json:"publickey"`
	Privatekey string `json:"privatekey"`
	Config     string `json:"config"`
	QRCode     string `json:"qrcode"`
	IsActive   bool   `json:"is_active"`
}

func (p *Profile) Validate() error {
	return validation.ValidateStruct(
		p,
		validation.Field(&p.Username, validation.Required, validation.Length(2, 255)),
		validation.Field(&p.Type, validation.Required, validation.Length(2, 255)),
	)
}

func (p *Profile) ReadKeys() (string, string, error) {
	file, err := os.Open(fmt.Sprintf("%spublickey", p.Path))
	if err != nil {
		return "", "", err
	}

	publickey, err := io.ReadAll(file)
	if err != nil {
		return "", "", err
	}
	file.Close()

	file, err = os.Open(fmt.Sprintf("%sprivatekey", p.Path))
	if err != nil {
		return "", "", err
	}

	privatekey, err := io.ReadAll(file)
	if err != nil {
		return "", "", err
	}
	file.Close()

	return string(publickey), string(privatekey), nil
}

func (p *Profile) AppendPear() error {
	f, err := os.OpenFile("/etc/wireguard/wgtest.conf", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("\n[Peer]\nPublicKey = %sAllowedIPs = 10.0.0.%d/32\n", p.Publickey, p.ID)); err != nil {
		return err
	}

	return nil
}

// GenProfileFiles <--
func (p *Profile) GenProfile() (string, error) {
	file, err := os.Open("/etc/wireguard/publickey")
	if err != nil {
		return "", err
	}

	publickey, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	file.Close()

	t := fmt.Sprintf("[Interface]\n"+
		"PrivateKey = %s\n"+
		"Address = 10.0.0.%d/32\n"+
		"DNS = 8.8.8.8\n\n"+
		"[Peer]\n"+
		"PublicKey = %s\n"+
		"Endpoint = %s:51880\n"+
		"AllowedIPs = 0.0.0.0/0\n"+
		"PersistentKeepalive = 20", p.Privatekey, p.ID, publickey, "0.0.0.0")

	f, err := os.OpenFile(fmt.Sprintf("%s/wg.conf", p.Path), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err = f.WriteString(t); err != nil {
		return "", err
	}

	path, _ := os.Getwd()

	if err := qrcode.WriteFile(t, qrcode.Medium, 512, fmt.Sprintf("%s/web/img/%d_wg.png", path, p.ID)); err != nil {
		return "", err
	}

	return t, nil
}

func (p *Profile) DelProfileFiles() error {
	if err := os.RemoveAll(p.Path); err != nil {
		return err
	}

	if err := os.Remove(fmt.Sprintf("/var/www/wg-admin/web/img/%d_wg.png", p.ID)); err != nil {
		return err
	}

	input, err := os.ReadFile("/etc/wireguard/wgtest.conf")
	if err != nil {
		return err
	}

	output := bytes.Replace(input, []byte(fmt.Sprintf("\n[Peer]\nPublicKey = %sAllowedIPs = 10.0.0.%d/32\n", p.Publickey, p.ID)), []byte(""), -1)

	err = os.WriteFile("/etc/wireguard/wgtest.conf", []byte(output), 0777)
	if err != nil {
		return err
	}

	return nil
}
