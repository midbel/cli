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
	"strings"
	"time"
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

type Size float64

const (
	Byte Size = 1
	Kilo      = 1024
	Mega      = Kilo * Kilo
	Giga      = Mega * Kilo
)

func ParseSize(v string) (Size, error) {
	var s Size
	if err := s.Set(v); err != nil {
		return 0, err
	}
	return s, nil
}

func (s *Size) Float() float64 {
	return float64(*s)
}

func (s *Size) Int() int64 {
	return int64(*s)
}

func (s *Size) String() string {
	var (
		u string
		v float64
	)
	switch s := *s; {
	case s < Kilo:
		u, v = "B", float64(s)
	case s >= Kilo && s < Mega:
		u, v = "KB", float64(s)/float64(Giga)
	case s >= Mega && s < Giga:
		u, v = "MB", float64(s)/float64(Mega)
	default:
		u, v = "GB", float64(s)/float64(Giga)
	}
	return fmt.Sprintf("%.2f%s", v, u)
}

func (s *Size) Divide(n int) Size {
	return Size(float64(*s) / float64(n))
}

func (s *Size) Multiply(n int) Size {
	return Size(float64(*s) * float64(n))
}

func (s *Size) Set(v string) error {
	var (
		f float64
		u string
	)
	n, err := fmt.Sscanf(v, "%f%s", &f, &u)
	if err != nil && n == 0 {
		return err
	}
	switch u {
	case "", "B":
	case "b":
		f /= 8
	case "KB", "K":
		f *= 1024
	case "kb", "k":
		f *= (1024 / 8)
	case "MB", "M":
		f *= 1024 * 1024
	case "mb", "m":
		f *= ((1024 * 1024) / 8)
	case "GB", "G":
		f *= 1024 * 1024 * 1024
	case "gb", "g":
		f *= ((1024 * 1024 * 1024) / 8)
	default:
		return fmt.Errorf("unknown unit given %s", u)
	}
	*s = Size(f)
	return nil
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
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
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

type Command struct {
	Desc  string
	Usage string
	Short string
	Alias []string
	Flag  flag.FlagSet
	Run   func(*Command, []string) error
}

func (c Command) Help() {
	if len(c.Desc) > 0 {
		fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Desc))
	} else {
		fmt.Fprintln(os.Stderr, c.Short)
	}
	fmt.Fprintf(os.Stderr, "\nusage: %s\n", c.Usage)
	os.Exit(2)
}

func (c Command) String() string {
	ix := strings.Index(c.Usage, " ")
	if ix < 0 {
		return c.Usage
	}
	return c.Usage[:ix]
}

func (c Command) Runnable() bool {
	return c.Run != nil
}
