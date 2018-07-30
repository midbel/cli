package cli

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	Version   string = ""
	BuildTime string = ""
)

const (
	BlockTypeRSA   = "RSA PRIVATE KEY"
	BlockTypeECDSA = "EC PRIVATE KEY"
	BlockTypeCert  = "CERTIFICATE"
	BlockTypeCSR   = "CERTIFICATE REQUEST"
)

type Time struct {
	time.Time
}

func (t *Time) String() string {
	return t.Time.String()
}

func (t *Time) Set(v string) error {
	if v == "" {
		t.Time = time.Now()
		return nil
	}
	i, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return err
	}
	t.Time = i
	return nil
}

type Certificate struct {
	Cert *x509.Certificate
}

func (c *Certificate) String() string {
	return "CERTIFICATE"
}

func (c *Certificate) Set(v string) error {
	bs, err := ioutil.ReadFile(v)
	if err != nil {
		return err
	}
	b, _ := pem.Decode(bs)
	cert, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		return err
	}
	c.Cert = cert
	return nil
}

type PrivateKey struct {
	Key crypto.PrivateKey
}

func (p *PrivateKey) String() string {
	return "PRIVATE KEY"
}

func (p *PrivateKey) Set(v string) error {
	bs, err := ioutil.ReadFile(v)
	if err != nil {
		return err
	}
	b, _ := pem.Decode(bs)

	var key crypto.Signer
	switch b.Type {
	case BlockTypeRSA:
		key, err = x509.ParsePKCS1PrivateKey(b.Bytes)
	case BlockTypeECDSA:
		key, err = x509.ParseECPrivateKey(b.Bytes)
	default:
		return fmt.Errorf("unsupported key type %s", b.Type)
	}
	p.Key = key
	return err
}

func IsDaemon() bool {
	if os.Getppid() != 1 {
		return false
	}
	for _, f := range []*os.File{os.Stdout, os.Stderr} {
		s, err := f.Stat()
		if err != nil {
			return false
		}
		m := s.Mode() & os.ModeDevice
		if m != 0 {
			return false
		}
	}
	return true
}

func Run(cs []*Command, usage func(), hook func(*Command) error) error {
	if hook == nil {
		hook = func(_ *Command) error {
			return nil
		}
	}
	version := struct {
		Short bool
		Long  bool
	}{}
	flag.BoolVar(&version.Short, "v", false, "")
	flag.BoolVar(&version.Long, "version", false, "")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if version.Short || version.Long || (len(args) > 0 && args[0] == "version") {
		printVersion()
		return nil
	}
	if len(args) == 0 || args[0] == "help" {
		flag.Usage()
		return nil
	}

	set := make(map[string]*Command)
	for _, c := range cs {
		if !c.Runnable() {
			continue
		}
		set[c.String()] = c
		for _, a := range c.Alias {
			set[a] = c
		}
	}
	if c, ok := set[args[0]]; ok && c.Runnable() {
		c.Flag.Usage = c.Help
		if err := hook(c); err != nil {
			return err
		}
		return c.Run(c, args[1:])
	}
	n := filepath.Base(os.Args[0])
	return fmt.Errorf(`%s: unknown subcommand "%s". run  "%[1]s help" for usage`, n, args[0])
}

func printVersion() {
	name, syst, arch := filepath.Base(os.Args[0]), runtime.GOOS, runtime.GOARCH
	if BuildTime == "" {
		t := time.Now()
		if p, err := os.Executable(); err == nil {
			if i, err := os.Stat(p); err == nil {
				t = i.ModTime().Truncate(time.Hour)
			}
		}
		BuildTime = t.Format(time.RFC3339)
	}
	if Version == "" {
		fmt.Printf("%s unknown %s/%s %s", name, syst, arch, BuildTime)
	} else {
		fmt.Printf("%s version %s %s/%s %s", name, Version, syst, arch, BuildTime)
	}
	fmt.Println()
}

type Command struct {
	Desc  string
	Usage string
	Short string
	Alias []string
	Flag  flag.FlagSet
	Run   func(*Command, []string) error
}

func (c *Command) Help() {
	if len(c.Desc) > 0 {
		fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Desc))
	} else {
		fmt.Fprintln(os.Stderr, c.Short)
	}
	fmt.Fprintf(os.Stderr, "\nusage: %s\n", c.Usage)
	os.Exit(2)
}

func (c *Command) String() string {
	ix := strings.Index(c.Usage, " ")
	if ix < 0 {
		return c.Usage
	}
	return c.Usage[:ix]
}

func (c *Command) Runnable() bool {
	return c.Run != nil
}
