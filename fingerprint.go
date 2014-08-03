// Package fingerprint provides functionality to calculate, compare and analyse
// acoustic fingerprints of raw audio data. According to Wikipedia, acoustic
// fingerprint is a condensed digital summary, deterministically generated
// from an audio signal, that can be used to identify an audio sample or
// quickly locate similar items in an audio database.
//
// Installation
//
// 	go get https://github.com/go-fingerprint/fingerprint
//
// You should also install any package containing the implementation of any
// fingerprinting algoritms. Currently only bindings to chromaprint library are
// supported.
//
// 	go get https://github.com/go-fingerprint/gochroma
//
// Usage
//
//	reader, _ := os.Open("test.raw")
// 	fpcalc := chromaprint.New(chromaprint.AlgorithmDefault)
// 	defer fpcalc.Close()
// 	fprint, _ := fpcalc.Fingerprint(
// 		fingerprint.RawInfo{
//			Src: reader,
// 			Channels: 2,
// 			Rate: 44100,
// 			MaxSeconds: 120,
// 		})
// 	// do anything with fingerprint...
package fingerprint

import (
	"errors"
	"image"
	"image/color"
	"io"
	"strconv"
	"strings"
)

// ErrLength describes a error that occurs when trying to compare fingerprints
// with different length.
var (
	ErrLength = errors.New(`fingerprint: unable to compare fingerprints with
	different length`)
)

const (
	bitsperint = 32
)

// Calculator is an interface type that can calculate
// acoustic fingerprints and return as raw int32 data
// or as base64-encoded string.
type Calculator interface {
	Fingerprint(i RawInfo) (fprint string, err error)
	RawFingerprint(i RawInfo) (fprint []int32, err error)
}

// A RawInfo holds information about raw audio data.
type RawInfo struct {
	// Reader connected with the audio data stream
	Src io.Reader
	// Number of channels of audio stream
	Channels uint
	// Sampling rate of input audio data, e.g. 44100
	Rate uint
	// Maximum number of seconds that will be taken
	// from the audio stream
	MaxSeconds uint
}

// Compare returns a number that indicates how two fingerprints
// are similar to each other as a value from 0 to 1. Usually two
// fingerprints can be considered identical when the score is
// greater or equal than 0.95.
func Compare(fprint1, fprint2 []int32) (float64, error) {
	dist := 0
	if len(fprint1) != len(fprint2) {
		return 0, ErrLength
	}

	for i, sub := range fprint1 {
		dist += hamming(sub, fprint2[i])
	}

	score := 1 - float64(dist)/float64(len(fprint1)*bitsperint)
	return score, nil
}

// Distance returns slice of pairwisely XOR-ed fingerprints.
func Distance(fprint1, fprint2 []int32) ([]int32, error) {
	if len(fprint1) != len(fprint2) {
		return nil, ErrLength
	}

	dist := make([]int32, len(fprint1))

	for i, sub := range fprint1 {
		dist[i] = sub ^ fprint2[i]
	}
	return dist, nil
}

// ToImage returns black-and-white image.Image with graphical
// representation of fingerprint: each column represents
// 32-bit integer, where black and white pixels correspond
// to 1 and 0 respectively.
func ToImage(fprint []int32) (im image.Image) {
	return int32ToImage(fprint)
}

// ImageDistance returns black-and white image.Image with
// graphical representation of distance between fingerprints.
func ImageDistance(fprint1, fprint2 []int32) (im image.Image, err error) {
	if len(fprint1) != len(fprint1) {
		return nil, ErrLength
	}

	dist, err := Distance(fprint1, fprint2)
	if err != nil {
		return
	}
	im = int32ToImage(dist)
	return
}

func hamming(a, b int32) (dist int) {
	dist = strings.Count(strconv.FormatInt(int64(a^b), 2), "1")
	return
}

func int32ToImage(s []int32) image.Image {
	im := image.NewGray(image.Rect(0, 0, len(s), bitsperint))
	for i, sub := range s {
		for j := 0; j < bitsperint; j++ {
			im.Set(i, j, color.Gray{uint8(sub&1) * 0xFF})
			sub >>= 1
		}
	}
	return im
}
