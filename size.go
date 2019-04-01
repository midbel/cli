package cli

import (
	"fmt"
)

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

func (s Size) Float() float64 {
	return float64(s)
}

func (s Size) Int() int64 {
	return int64(s)
}

func (s Size) String() string {
	return formatSize(float64(s))
}

func (s Size) Divide(n int) Size {
	return Size(float64(s) / float64(n))
}

func (s Size) Multiply(n int) Size {
	return Size(float64(s) * float64(n))
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

func formatSize(s float64) string {
	var (
		u string
		v float64
	)
	switch {
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
