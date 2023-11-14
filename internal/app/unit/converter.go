// Package converter implements function to convert units
package converter

const (
	unit          = 1024
	secondsPerDay = 86400
)

type Number interface {
	int32 | int64 | float64
}

func GigaBytesToBytes[N Number](size N) N {
	return size * unit * unit * unit
}

func MegaBytesToBytes[N Number](size N) N {
	return size * unit * unit
}

func KiloByteToMegaBytes[N Number](size N) N {
	return size / unit
}

func DaystoSeconds[N Number](days N) N {
	return days * secondsPerDay
}
