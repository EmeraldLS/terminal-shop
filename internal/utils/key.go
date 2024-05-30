package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

func GetPublicKey(userHostAndPort string) (string, string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", "", err
	}

	authorizedKeysPath := filepath.Join(currentUser.HomeDir, ".ssh", "known_hosts")
	authorizedKeysBytes, err := os.ReadFile(authorizedKeysPath)
	if err != nil {
		return "", "", err
	}

	authorizedKeys := string(authorizedKeysBytes)

	keyLines := strings.Split(authorizedKeys, "\n")

	for i, line := range keyLines {
		if len(keyLines[i]) > 2 {

			lineArr := strings.Split(line, " ")
			hostAndPort := strings.Split(lineArr[0], ":")
			hostStr := hostAndPort[0]
			var sb strings.Builder
			var newHosts []string
			for _, h := range hostStr {
				if h == ' ' {
					newHosts = append(newHosts, sb.String())
					sb.Reset()
				} else if h != '[' && h != ']' {
					_, err := sb.WriteRune(h)
					if err != nil {
						return "", "", err

					}
				}
			}

			if sb.Len() > 0 {
				newHosts = append(newHosts, sb.String())
			}

			port := hostAndPort[1]

			userHostArr := strings.Split(userHostAndPort, ":")
			if len(userHostArr) < 2 {
				return "", "", errors.New("inavlid Input for Host & port. Expect e.g localhost:<port>")
			}

			userHost := userHostArr[0]
			userPort := userHostArr[1]

			for _, host := range newHosts {
				if host == userHost && port == userPort {
					algoIdentifierGotten := lineArr[1]
					pubkeyGotten := lineArr[2]
					return algoIdentifierGotten, pubkeyGotten, nil
				}
			}

		}
	}

	return "", "", errors.New("unable to get public keys. Invalid host or port provided")
}

func MakeSSHKeyPair() (string, string, error) {
	// generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", "", err
	}

	// write private key as PEM
	var privKeyBuf strings.Builder

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(&privKeyBuf, privateKeyPEM); err != nil {
		return "", "", err
	}

	// generate and write public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	var pubKeyBuf strings.Builder
	pubKeyBuf.Write(ssh.MarshalAuthorizedKey(pub))

	return pubKeyBuf.String(), privKeyBuf.String(), nil
}
