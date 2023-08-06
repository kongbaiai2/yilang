package utils

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"time"

	"golang.org/x/crypto/ssh"
)

type SshCli struct {
	User   string
	Pwd    string
	Addr   string // ip:port
	client *ssh.Client

	LastResult     string
	useKey         bool
	PrivateKeyPath string
	PrivateKey     string
}

func (c *SshCli) Connect() (*SshCli, error) {
	config := &ssh.ClientConfig{Timeout: 30 * time.Second}
	config.SetDefaults()
	config.User = c.User
	config.Auth = []ssh.AuthMethod{ssh.Password(c.Pwd)}
	config.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }
	client, err := ssh.Dial("tcp", c.Addr, config)
	if nil != err {
		return c, err
	}
	c.client = client
	return c, nil
}

func (c *SshCli) PublicyKeyConnect() (*SshCli, error) {

	var key []byte
	// transport key file path or get user default key file path
	if c.PrivateKeyPath == "" {
		homePath, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		key, err = ioutil.ReadFile(path.Join(homePath, ".ssh", c.PrivateKey))
		if err != nil {
			return nil, err
		}
	} else {
		exist, err := PathExists(c.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		if exist {
			key, err = ioutil.ReadFile(c.PrivateKeyPath)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("private key not found: %s", c.PrivateKeyPath)
		}

	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", c.Addr, config)
	if nil != err {
		log.Println(c, config)
		return c, err
	}
	c.client = client
	return c, nil
}

func (c *SshCli) RunShell(cmds []string, inChan chan string) error {
	defer close(inChan)
	if c.useKey {
		if _, err := c.PublicyKeyConnect(); err != nil {
			return err
		}
		defer c.client.Close()
	} else {
		if _, err := c.Connect(); err != nil {
			return err
		}
		defer c.client.Close()
	}

	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	// fd := int(os.Stdin.Fd())

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	// term := os.Getenv("TERM")
	// if term == "" {
	// 	term = "xterm"
	// }

	// err = session.RequestPty(term, 80, 40, modes)
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Fatalf("request for pseudo terminal failed: %v", err)
		return err
	}

	session.Stderr = os.Stderr
	// var out bytes.Buffer
	// session.Stdout = &out

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatal("输入错误", err)
		return err
	}

	cmdReader, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(cmdReader)
	go func(inChan chan string) {
		for scanner.Scan() {
			if scanner.Text() == "Unrecognized command found" ||
				scanner.Text() == "Too many parameters found" {
				return
			}
			inChan <- scanner.Text()
		}
	}(inChan)

	if err = session.Shell(); err != nil {
		log.Fatal("创建shell出错", err)
		return err
	}

	_, _ = fmt.Fprintf(stdin, "%s\n", "screen-length 512 temporary")
	_, _ = fmt.Fprintf(stdin, "%s\n", "sys")
	for _, cmd := range cmds {
		// log.Println(cmd)
		_, err = fmt.Fprintf(stdin, "%s\n", cmd)
		if err != nil {
			log.Fatal("写入stdin出错", err)
			return err
		}
	}
	_, _ = fmt.Fprintf(stdin, "%s\n", "ret")
	_, _ = fmt.Fprintf(stdin, "%s\n", "quit")

	// log.Print("result: ", strings.Join(result, "\n"))

	// time.Sleep(10 * time.Second)

	if err = session.Wait(); err != nil {
		return err
	}
	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GenerateKey(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	private, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return private, &private.PublicKey, nil

}

func EncodePrivateKey(private *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(private),
		Type:  "RSA PRIVATE KEY",
	})
}

func EncodePublicKey(public *rsa.PublicKey) ([]byte, error) {
	publicBytes, err := x509.MarshalPKIXPublicKey(public)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Bytes: publicBytes,
		Type:  "PUBLIC KEY",
	}), nil
}

// EncodeSSHKey
func EncodeSSHKey(public *rsa.PublicKey) ([]byte, error) {
	publicKey, err := ssh.NewPublicKey(public)
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(publicKey), nil
}

func MakeSSHKeyPair(homePath, pubname, priname string) error {
	pkey, pubkey, err := GenerateKey(1024)
	if err != nil {
		return err
	}
	pub, err := EncodeSSHKey(pubkey)
	if err != nil {
		return err
	}

	ioutil.WriteFile(pubname, pub, 0644)
	ioutil.WriteFile(priname, EncodePrivateKey(pkey), 0600)

	return nil
}

// func (wsw *wsWrapper) Write(p []byte) (n int, err error) {
// 	writer, err := wsw.Conn.NextWriter(websocket.TextMessage)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer writer.Close()
// 	return writer.Write(p)
// }

// func (wsw *wsWrapper) Read(p []byte) (n int, err error) {
// 	for {
// 		msgType, reader, err := wsw.Conn.NextReader()
// 		if err != nil {
// 			return 0, err
// 		}
// 		if msgType != websocket.TextMessage {
// 			continue
// 		}
// 		return reader.Read(p)
// 	}
// }

func exists(p string) (bool, os.FileInfo) {
	f, err := os.Open(p)
	if err != nil {
		return false, nil
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return false, nil
	}
	return true, fi
}

func isDir(p string) bool {
	if e, fi := exists(p); e {
		return fi.Mode().IsDir()
	}
	err := os.MkdirAll(p, os.ModePerm)
	if err != nil {
		return false
	}
	if ee, fii := exists(p); ee {
		return fii.Mode().IsDir()
	}
	return false
}

func NewSshCfg() *SshCli {
	return &SshCli{}
}
